package controllers

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	udomain "energyshares-invest-worker/domain"
)

// SQSワーカーを構築するための構造体
type SqsWorker struct {
	sqsClient     *sqs.Client
	cfg           config.WorkerConfig
	queueURL      string
	investUsecase udomain.InvestUsecase
	shutdownCh    chan struct{}
	wg            sync.WaitGroup
}

// NewSqsWorker は SqsWorker を初期化して返すコンストラクタ。
//
// SQS クライアント・ワーカー設定・キュー URL・投資ユースケースを受け取り、
// シャットダウン用チャネルを生成した状態で返す。
func NewSqsWorker(sqsClient *sqs.Client, workerConfig config.WorkerConfig, queueURL string, investUsecase udomain.InvestUsecase) *SqsWorker {
	return &SqsWorker{
		sqsClient:     sqsClient,
		cfg:           workerConfig,
		queueURL:      queueURL,
		investUsecase: investUsecase,
		shutdownCh:    make(chan struct{}),
	}
}

// StartPolling は SQS キューへのロングポーリングを開始するメソッド。
//
// context がキャンセルされるか Shutdown が呼ばれるまでポーリングを継続する。
// メッセージ受信エラー時は指数バックオフでリトライし、
// 最大リトライ数（MaxRetryCount）を超えた場合は panic でワーカーを停止する。
// メッセージを受信した場合は processMessage を呼び出して処理する。
func (w *SqsWorker) StartPolling(ctx context.Context) {
	w.wg.Add(1)
	defer w.wg.Done()

	retryCount := int32(0)
	for {
		// ループ終了条件のチェック
		select {
		case <-ctx.Done():
			logrus.Info("Polling worker is shutting down.")
			return
		case <-w.shutdownCh:
			logrus.Info("Shutdown signal received, stopping polling")
			return
		default:
			// ポーリング処理を続行
		}

		// SQSからメッセージを受信（ロングポーリングを使用）
		receiveResult, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(w.queueURL),
			MaxNumberOfMessages: w.cfg.MaxNumberOfSqsMessages,
			WaitTimeSeconds:     w.cfg.WaitTimeSecondsForSqsMessages,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to receive messages from SQS.")

			// 指数バックオフの実装
			if retryCount >= w.cfg.MaxRetryCount {
				logrus.Errorf("Exceeded max retries (%d). Shutting down worker.", w.cfg.MaxRetryCount)
				panic("SQS worker failed to receive messages after multiple retries.")
			}
			backoffDuration := time.Duration(math.Pow(2, float64(retryCount))) * time.Second
			logrus.Infof("Retrying in %s...", backoffDuration)
			time.Sleep(backoffDuration)
			retryCount++
			continue
		}

		retryCount = 0

		if len(receiveResult.Messages) == 0 {
			logrus.Info("No messages in queue, waiting...")
			continue
		}

		// 受信した各メッセージを処理
		for _, message := range receiveResult.Messages {
			if err := w.processMessage(ctx, message); err != nil {
				logrus.WithError(err).Error("Failed to process message.")
			}
		}
	}
}

// processMessage は SQS から受信した1件のメッセージを処理する内部メソッド。
//
// 以下の順序で処理を行う:
//  1. シャットダウン信号を受けていた場合は即座に処理をスキップ
//  2. メッセージボディが nil の場合はキューから削除してエラーを返す
//  3. ボディを InvestRequest にアンマーシャル。失敗時はキューから削除してエラーを返す
//  4. InvestUsecase.ProcessInvestmentRequest を呼び出して投資処理を実行
//  5. 処理成功後、キューからメッセージを削除
func (w *SqsWorker) processMessage(ctx context.Context, message types.Message) error {
	select {
	case <-w.shutdownCh:
		logrus.Info("Shutdown requested, skipping message processing")
		return nil
	default:
	}

	if message.Body == nil {
		_, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(w.queueURL),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			return errors.Wrap(err, "failed to delete message")
		}
		return errors.Wrap(err, "Empty message body - deleted message")
	}

	var req udomain.InvestRequest
	err := json.Unmarshal([]byte(*message.Body), &req)
	if err != nil {
		// Jsonのパースに失敗した場合、ログを出力してメッセージを削除
		_, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(w.queueURL),
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			return errors.Wrap(err, "failed to delete message")
		}
		return errors.Wrap(err, "failed to unmarshal message body - deleted message")
	}

	if err := w.investUsecase.ProcessInvestmentRequest(ctx, req); err != nil {
		return errors.Wrap(err, "fail investUsecase ProcessInvestmentRequest()")
	}

	// メッセージ削除
	_, err = w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(w.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete message")
	}

	return nil
}

// Shutdown は SqsWorker をグレースフルにシャットダウンするメソッド。
//
// shutdownCh をクローズして StartPolling および processMessage にシャットダウン信号を送り、
// 実行中の全ワーカーゴルーチンが完了するまで待機する。
// context のタイムアウトを超えた場合はエラーを返す。
func (w *SqsWorker) Shutdown(ctx context.Context) error {
	logrus.Info("Initiating graceful shutdown of SQS worker...")

	// shutdown信号を送信
	close(w.shutdownCh)

	// contextのタイムアウトまで待機
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logrus.Info("All workers stopped successfully")
		return nil
	case <-ctx.Done():
		return errors.New("shutdown timeout exceeded")
	}
}
