package server

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSkipIfStillRunningSkipsConcurrentInvocation(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var runCount int32

	wrapped := skipIfStillRunning(lotteryCSVUploadJobName, func(ctx context.Context) error {
		atomic.AddInt32(&runCount, 1)
		close(started)
		<-release
		return nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = wrapped(context.Background())
	}()

	<-started

	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("wrapped() error = %v", err)
	}

	close(release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("first invocation did not finish in time")
	}

	if got := atomic.LoadInt32(&runCount); got != 1 {
		t.Fatalf("runCount = %d, want 1", got)
	}
}

func TestJobNeedsSkipIfStillRunning(t *testing.T) {
	tests := []struct {
		jobName string
		want    bool
	}{
		{lotteryCSVUploadJobName, true},
		{rePublishJobName, true},
		{"MailRequestRetryJob", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.jobName, func(t *testing.T) {
			if got := jobNeedsSkipIfStillRunning(tt.jobName); got != tt.want {
				t.Fatalf("jobNeedsSkipIfStillRunning(%q) = %v, want %v", tt.jobName, got, tt.want)
			}
		})
	}
}

// MIR-1470: JPKI方式の署名結果を拾い直すジョブは、前回の実行が終わっていない間は次の実行をスキップする。
func TestJobNeedsSkipIfStillRunning_JpkiSignatureRecovery(t *testing.T) {
	assert.True(t, jobNeedsSkipIfStillRunning(jpkiSignatureRecoveryJobName))
	assert.True(t, jobNeedsSkipIfStillRunning(lotteryCSVUploadJobName), "既存のジョブは変わらない")
	assert.False(t, jobNeedsSkipIfStillRunning("RefreshTokenJob"), "対象でないジョブは保護しない")
}

// 設定(cron)に JpkiSignatureRecoveryJob が無いと、ジョブは登録されず、エラーにもならない。
// 有ると、その式で登録される(大文字小文字は区別しない)。
func TestGetCronExpr_JpkiSignatureRecovery(t *testing.T) {
	assert.Equal(t, "7 * * * *", getCronExpr(map[string]string{"JpkiSignatureRecoveryJob": "7 * * * *"}, jpkiSignatureRecoveryJobName))
	assert.Equal(t, "7 * * * *", getCronExpr(map[string]string{"jpkisignaturerecoveryjob": "7 * * * *"}, jpkiSignatureRecoveryJobName))
	assert.Empty(t, getCronExpr(map[string]string{}, jpkiSignatureRecoveryJobName), "設定が無ければ登録しない")
	assert.Empty(t, getCronExpr(map[string]string{"JpkiSignatureRecoveryJob": "  "}, jpkiSignatureRecoveryJobName), "空の設定も登録しない")
}
