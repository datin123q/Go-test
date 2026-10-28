package controllers

import (
	"encoding/csv"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/csvutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/i18nutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
)

type MailApiController struct{}

func (c MailApiController) Init(g echoswagger.ApiGroup) {

	g.GET("/csv-form-download", c.GetCsvFormDownload).
		AddResponse(http.StatusOK, "受信者CSVファイル形式ダウンロード", nil, nil).
		SetDescription("受信者CSVファイル形式をダウンロードします.").
		SetSummary("受信者CSVファイル形式ダウンロード").
		SetSecurity("Authorization")

	g.POST("/send", c.SendMail).
		AddParamBody(MailSendOptionRequest{}, "body", "", true).
		AddResponse(http.StatusCreated, "メール送信成功の有無(true|false)", nil, nil).
		SetDescription("オプションに該当するユーザーにメールを送信します.").
		SetSummary("オプションを通じたメール送信").
		SetSecurity("Authorization")

	// フロントからのGETリクエストの場合bodyを使用できないのでPOSTを使用
	g.POST("/receivers/count", c.CountReceiversWithOption).
		AddParamBody(MailSendOptionRequest{}, "body", "", true).
		AddResponse(http.StatusCreated, "メール受信者数", CountReceiversWithOptionResponse{}, nil).
		SetDescription("オプションに該当するユーザーの数を照会します.").
		SetSummary("オプションに該当するユーザー数").
		SetSecurity("Authorization")

	// フロントからのGETリクエストの場合formを推奨しないのでPOSTを使用
	g.POST("/receivers/count-csv", c.CountReceiversFromCsv).
		AddParamFile("receiver_list", "受信者リストcsv ファイル", true).
		AddResponse(http.StatusCreated, "メール受信者数", CountReceiversFromCsv{}, nil).
		SetDescription("csvに記載された数を照会します.").
		SetSummary("csvに記載されたユーザー数").
		SetSecurity("Authorization")

	g.POST("/send/csv", c.SendMailByCsv).
		AddParamForm(MailSendCsvRequest{}, "body", "", true).
		AddParamFile("receiver_list", "受信者リストcsv ファイル", true).
		AddResponse(http.StatusCreated, "メール送信成功の有無 (true|false)", nil, nil).
		SetDescription("csvのアドレスリストにメールを送信します.").
		SetSummary("csvメール送信").
		SetSecurity("Authorization")

	g.GET("/receivers", c.GetAllReceivers).
		AddParamQueryNested(SearchInput{}).
		AddResponse(http.StatusOK, "メール受信者リスト照会", []MailReceiversResponse{}, nil).
		SetDescription("メール受信者リスト照会, (デフォルト条件: sortBy=email, order=asc, offset=0, limit=1000)").
		SetSummary("メール受信者リスト照会").
		SetSecurity("Authorization")

	g.GET("/senders", c.GetAllSenders).
		AddResponse(http.StatusOK, "メール送信者リスト照会", []string{}, nil).
		SetDescription("メール送信者リストを照会します.").
		SetSummary("メール送信者リスト照会").
		SetSecurity("Authorization")
}

func (MailApiController) SendMail(c echo.Context) error {

	var v MailSendOptionRequest
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	receivers := make([]string, 0)
	emailMap := make(map[string]struct{})
	options := v.ToDomain()

	// 受信者リスト照会
	if options.IsNotEmpty() {
		users, err := domain.Mail{}.GetReceivers(c.Request().Context(), options)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		for _, u := range users {
			if _, exists := emailMap[u.Email]; !exists {
				emailMap[u.Email] = struct{}{}
				receivers = append(receivers, u.Email)
			}
		}
	}

	// 直接入力した受信者追加
	for _, r := range v.Receiver {
		if _, exists := emailMap[r]; !exists {
			emailMap[r] = struct{}{}
			receivers = append(receivers, r)
		}
	}

	// 受信者がない場合error
	if len(receivers) <= 0 {
		return factory.ErrorNoRecipients.New(nil)
	}

	// メール送信
	if err := (domain.Mail{}).NormalEmail(c.Request().Context(), v.Sender, v.Subject, v.Content, receivers); err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	// audit log
	loginId, loginEmail, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), loginId, loginEmail,
		mailSendAuditMenu, truncateMailAuditTitle(v.Subject), c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, true)
}

func (MailApiController) CountReceiversWithOption(c echo.Context) error {

	var v MailSendOptionRequest
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	// countではタイトル、送信者などのrequired条件チェックなし
	// クリーンアーキテクチャでは別のdtoを使用するよう修正が必要
	// if err := c.Validate(&v); err != nil {
	//	return factory.ErrorParameter.New(err)
	// }

	receivers := make([]string, 0)
	emailMap := make(map[string]struct{})
	options := v.ToDomain()

	// 受信者リスト照会
	if options.IsNotEmpty() {
		users, err := domain.Mail{}.GetReceivers(c.Request().Context(), options)
		if err != nil {
			return factory.ErrorDB.New(err)
		}

		for _, u := range users {
			if _, exists := emailMap[u.Email]; !exists {
				emailMap[u.Email] = struct{}{}
				receivers = append(receivers, u.Email)
			}
		}
	}

	// 直接入力した受信者追加
	for _, r := range v.Receiver {
		if _, exists := emailMap[r]; !exists {
			emailMap[r] = struct{}{}
			receivers = append(receivers, r)
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, CountReceiversWithOptionResponse{Count: len(receivers)})
}

func (MailApiController) SendMailByCsv(c echo.Context) error {

	var v MailSendCsvRequest
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	receivers := make([]string, 0)
	csvDataList := make(map[string]bool)

	// csvから受信者リスト照会
	formFile, err := c.FormFile("receiver_list")
	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}

	file, err := formFile.Open()
	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}
	defer file.Close()

	lines, err := csvutils.ReadMultpartCsv(&file)
	if err != nil {
		return factory.ErrorOutOfCsvForm.New(err)
	}

	for index, record := range lines {
		if index < 5 {
			continue
		}

		_, ok := csvDataList[record[0]]
		if ok {
			continue
		} else {
			csvDataList[record[0]] = true
			receivers = append(receivers, record[0])
		}
	}

	// 受信者がない場合error
	if len(receivers) <= 0 {
		return factory.ErrorNoRecipients.New(nil)
	}

	// メール送信
	if err := (domain.Mail{}).NormalEmail(c.Request().Context(), v.Sender, v.Subject, v.Content, receivers); err != nil {
		return factory.ErrorAWSSendMail.New(err)
	}

	// audit log
	loginId, loginEmail, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), loginId, loginEmail,
		mailSendAuditMenu, truncateMailAuditTitle(v.Subject), c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, true)
}

func (MailApiController) CountReceiversFromCsv(c echo.Context) error {
	csvDataList := make(map[string]bool)
	receiverCount := 0

	// csvから受信者リスト照会
	formFile, err := c.FormFile("receiver_list")
	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}

	file, err := formFile.Open()
	if err != nil {
		return factory.ErrorNotSupportedFiles.New(err)
	}
	defer file.Close()

	lines, err := csvutils.ReadMultpartCsv(&file)
	if err != nil {
		return factory.ErrorOutOfCsvForm.New(err)
	}

	for index, record := range lines {
		if index < 5 {
			continue
		}

		_, ok := csvDataList[record[0]]
		if ok {
			continue
		} else {
			csvDataList[record[0]] = true
			receiverCount++
		}
	}

	return apirender.RenderSucc(c, http.StatusOK, CountReceiversFromCsv{Count: receiverCount})
}

func (MailApiController) GetAllReceivers(c echo.Context) error {
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = 999999
	}

	total, receivers, err := domain.User{}.GetAllEmailsAndNames(c.Request().Context(), &v.Order, &v.Sortby, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      MailReceiversResponse{}.ToDtoList(receivers),
		TotalCount: total,
	})
}

// GetAllSenders 発信者メールアドレスリスト
func (MailApiController) GetAllSenders(c echo.Context) error {
	items := serviceutils.GetServiceEmails()

	return apirender.RenderSucc(c, http.StatusOK, items)
}

// GetCsvFormDownload CSVファイル形式ダウンロード
func (MailApiController) GetCsvFormDownload(c echo.Context) error {

	records := [][]string{}
	records = append(records, []string{i18nutils.DefaultTranslate("model.csv_form.first_line_phrase")})
	records = append(records, []string{i18nutils.DefaultTranslate("model.csv_form.second_line_phrase")})
	records = append(records, []string{i18nutils.DefaultTranslate("model.csv_form.third_line_phrase")})
	records = append(records, []string{i18nutils.DefaultTranslate("model.csv_form.fourth_line_phrase")})
	records = append(records, []string{i18nutils.DefaultTranslate("model.csv_form.fifth_line_phrase")})

	w := c.Response().Writer

	csvutils.WriteHeader(w, "mail_form.csv")

	return csv.NewWriter(w).WriteAll(records)
}

// mailSendAuditMenu は一斉メール送信の監査ログ(OperatorHistory)に記録するmenuの値。
// SendMail・SendMailByCsvの2箇所で共有する。
const mailSendAuditMenu = "menu.manage_mail_sms.manage_emails_manage_send"

// mailAuditTitleMaxLen は operator_history.title(varchar(255))が受け付ける文字数の上限。
const mailAuditTitleMaxLen = 255

// truncateMailAuditTitle は監査ログに記録する件名をvarchar(255)に収まる文字数まで切り
// 詰める。MIR-1236: これまでは常に空文字を記録していたため顕在化しなかったが、実際の
// 件名を記録するようにしたことで、255文字を超える件名だとINSERTがエラーになる。
// admin-apiはリクエスト単位でDB書き込みを1つのトランザクションに包んでおり、このエラー
// でそれまでのDB書き込みが巻き戻される一方、メール送信(統合メールサービスへのHTTP
// リクエスト、domain.Mail.NormalEmail)はトランザクションの外の副作用のため取り消せない。
// 担当者が画面のエラーを送信失敗と誤解して再送すると二重送信になるため、記録側だけを
// 切り詰めて防ぐ(送信するメールの件名自体はそのまま)。根本原因(メール送信がDB
// トランザクションの外にあること)はMIR-1269のスコープ。
// varchar(255)はutf8mb4での文字数(コードポイント数)の上限であり、バイト数ではないため、
// []runeに変換して数える(バイト数で切ると絵文字等の途中で文字列が壊れる)。
func truncateMailAuditTitle(title string) string {
	r := []rune(title)
	if len(r) <= mailAuditTitleMaxLen {
		return title
	}
	return string(r[:mailAuditTitleMaxLen])
}
