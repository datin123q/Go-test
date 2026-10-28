package controllers

import (
	"context"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/illuminarean/energy-share-core-api/service"
	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/apirender"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/formutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jwthelper"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/s3utils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ProjectApiController struct {
	contentImageService *service.ContentImageService
}

func (c ProjectApiController) Init(g echoswagger.ApiGroup) {
	c.contentImageService = service.NewContentImageService()

	g.GET("", c.GetAll).
		AddParamQueryNested(ProjectSearchInput{}).
		AddResponse(http.StatusOK, "", ProjectAdminResponseDto{}, nil).
		SetDescription("ファンド一覧を取得します。").
		SetSummary("ファンド一覧取得").
		SetSecurity("Authorization")

	g.POST("", c.Create).
		AddParamForm(ProjectInput{}, "body", "", true).
		AddParamFile("main_image_file", "メイン画像", false).
		AddParamFile("sub_image_files", "サブ画像", false).
		AddResponse(http.StatusOK, "", ProjectAdminResponseDto{}, nil).
		SetDescription("ファンドを追加します。").
		SetSummary("ファンド追加").
		SetSecurity("Authorization")

	// 静的パスのため /:id より優先して解決される。新規登録時は id を持たないので、
	// パスパラメータではなくクエリで受け取る。
	g.GET("/check-fund-number", c.CheckDuplicateFundNumber).
		AddParamQuery("", "fund_number", "ファンド番号", true).
		AddParamQuery("", "id", "自身のファンドID。新規登録時は指定しない", false).
		AddResponse(http.StatusOK, "", CheckDuplicateFundNumberResponseDto{}, nil).
		SetDescription("ファンド番号が他のファンドで使われていないかを確認します。").
		SetSummary("ファンド番号重複確認").
		SetSecurity("Authorization")

	g.GET("/:id", c.GetOne).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "", ProjectAdminResponseDto{}, nil).
		SetDescription("ファンドの詳細を取得します。").
		SetSummary("ファンド詳細取得").
		SetSecurity("Authorization")

	// casbin_rule.csv 削除作業。PUT は存在せず、POST のみだったため、正常に動作していなかった可能性がある。
	g.PUT("/:id", c.Update).
		AddParamForm(ProjectInput{}, "body", "", true).
		AddParamPath("", "id", "").
		AddParamFile("main_image_file", "", false).
		AddParamFile("sub_image_files", "", false).
		AddResponse(http.StatusOK, "", ProjectAdminResponseDto{}, nil).
		SetDescription("ファンドを更新します。").
		SetSummary("ファンド更新").
		SetSecurity("Authorization")

	g.PUT("/:id/project-status/:project_status", c.UpdateProjectStatus).
		AddParamPath("", "id", "").
		AddParamPath(
			"",
			"project_status",
			"プロジェクトステータス入力 -> STANDBY_FUNDING, OPEN_FUNDING, END_FUNDING, SUCCESS_FUNDING, FAIL_FUNDING, OPERATION_FUNDING, OPERATION_END_FUNDING, CLOSE_FUNDING, INVALID_FUNDING",
		).
		AddParamQuery("", "confirm_remaining_principal", "償還完了にするとき、元本が残っている・分配・償還の予定が無いことへの確認に同意した場合 true", false).
		AddResponse(http.StatusOK, "プロジェクトステータス変更成功", &domain.Project{}, nil).
		AddResponse(http.StatusBadRequest, "プロジェクトステータス変更失敗", nil, nil).
		AddResponse(http.StatusConflict, "状態が変わっていた、または償還完了の確認が必要", nil, nil).
		SetSummary("ファンドステータス変更処理").
		SetDescription("ファンドステータスを変更します。").
		SetSecurity("Authorization")

	g.GET("/:id/close-check", c.GetCloseCheck).
		AddParamPath("", "id", "").
		AddResponse(http.StatusOK, "", domain.CloseCheck{}, nil).
		SetSummary("償還完了の確認").
		SetDescription("償還完了にする前に、未実施の予定の件数・最後の予定の元本残高・確認が要るか(confirmation_required)を返します。画面は確認の要否をこれで決めます(MIR-272)。").
		SetSecurity("Authorization")

	// casbin_rule.csv 削除作業。現在使用中だが casbin_rule.csv には存在していなかった。
	g.DELETE("/:id", c.Delete).
		AddParamPath("", "id", "").
		SetDescription("該当ファンドを削除します。").
		SetSummary("ファンド削除").
		SetSecurity("Authorization")

	g.GET("/list/download", c.Download).
		AddParamQueryNested(ProjectSearchInput{}).
		AddResponse(http.StatusOK, "", domain.Project{}, nil).
		SetDescription("ファンド一覧をダウンロードします。").
		SetSummary("ファンド一覧ダウンロード").
		SetSecurity("Authorization")

	g.GET("/:id/investments", c.GetInvestmentsByProjectId).
		SetSummary("ファンド投資取引履歴取得").
		AddParamPath("", "id", "").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "商品に関連するすべての投資取引履歴情報", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("商品に関連するすべての投資取引履歴を取得します。")

	g.GET("/:id/investments/download", c.InvestmentsDownload).
		SetSummary("ファンド投資取引履歴ダウンロード").
		AddParamPath("", "id", "").
		AddParamQuery(SearchInput{}, "", "", true).
		AddResponse(http.StatusOK, "商品に関連するすべての投資取引履歴情報", domain.Investment{}, nil).
		SetSecurity("Authorization").
		SetDescription("ファンドに関連するすべての投資取引履歴をダウンロードします。")
}

func (ProjectApiController) GetAll(c echo.Context) error {
	var v ProjectSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	uid, uemail, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"user":             uid,
		"email":            uemail,
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")
	var totalCount int64
	var items []domain.Project

	totalCount, items, err = domain.Project{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount, v.ProjectId, v.ProjectStatus, v.TargetShareAmount, v.ProjectName, v.ProjectType, v.CategoryType, v.BlockInvest, v.StartDate, v.EndDate, v.OutOfDateProject, v.Posted)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})
}

func (p ProjectApiController) Create(c echo.Context) error {
	var v ProjectInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	tokenId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return factory.ErrorInvalidForm.New(err)
	}

	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	project, err := v.ToModel()
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// 重複の判定は INSERT より前に行う。fund_number にはユニークインデックスが
	// あるため、後ろに置くと INSERT が先に落ちて「DB error」になり、
	// ここまで到達しない。
	//
	// closed_fund_slug には制約が無く INSERT は通ってしまうが、後ろに置くと
	// ファンドを作って投資家へメールを送ったあとに弾くことになる。
	//
	// 新規登録時は id が 0 のため、id != 0 の条件で全件が比較対象になる。
	exist, err := project.CheckExistClosedSlug(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if exist {
		return factory.ErrorAlreadyExists.New(errors.New("project with the same closed slug already exists"))
	}

	existFundNumber, err := project.CheckExistFundNumber(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if existFundNumber {
		return factory.ErrorAlreadyExists.New(errors.New("project with the same fund number already exists"))
	}

	if _, err := project.Create(c.Request().Context(), tokenId); err != nil {
		return factory.ErrorDB.New(err)
	}

	if err := p.projectContentHtmlProcess(c, project, form); err != nil {
		return factory.ErrorDB.New(errors.New("failed project content html process"))
	}

	receivers, err := p.GetProjectInvestorMails(c, project)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	newsInputs, err := p.createNewsInputs(c, form, project)
	if err != nil {
		return err
	}

	err = p.createProjectNews(c, newsInputs, nil, receivers, project, tokenId)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	progressItemInputs, err := p.createMilestoneInputs(form)
	if err != nil {
		return err
	}
	err = p.createMilestone(c, progressItemInputs, project)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	_, mImgFile, _ := c.Request().FormFile("main_image_file")
	if err := p.mainImageUpload(c, tokenId, mImgFile, project); err != nil {
		return factory.ErrorAWSUpload.New(errors.New("failed project main image uploads"))
	}

	subImages := map[string]*multipart.FileHeader{}
	for name, value := range form.File {
		if strings.Contains(name, "sub_image_files") && len(value) > 0 {
			if index, ok := extractSubImageIndex(name); ok {
				subImages[index] = value[0]
			}
		}
	}

	subImageOrder := map[string]int64{}
	for name, value := range form.Value {
		if strings.Contains(name, "sub_image_order") && len(value) > 0 {
			if index, ok := extractSubImageIndex(name); ok {
				order, err := strconv.Atoi(value[0])
				if err != nil {
					return factory.ErrorParameter.New(err)
				}
				subImageOrder[index] = int64(order)
			}
		}
	}

	if len(subImages) > 0 {
		if err := p.uploadSubImage(c, tokenId, subImages, project, subImageOrder); err != nil {
			return factory.ErrorAWSUpload.New(errors.New("failed project sub image uploads"))
		}
	}

	updatedProjectImages, err := domain.ProjectImage{}.FindByProjectIdAndImageType(c.Request().Context(), project.Id, domain.SUB_IMG)
	if err != nil {
		return factory.ErrorDB.New(err, "failed get project main images")
	}
	for _, img := range updatedProjectImages {
		project.SubImageInfoList = append(project.SubImageInfoList, &domain.SubImageInfo{
			SubImageId:  img.Id,
			SubImageUrl: img.FileUrl,
			SubImage:    img,
		})
	}

	err = p.UploadProjectFile(c, form, project)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if err := project.Update(c.Request().Context()); err != nil {
		return factory.ErrorDB.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), tokenId, email,
		"menu.manage_product.product_register", project.Name, c.RealIP(), domain.REGISTER)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, ProjectAdminResponseDto{}.New(project))
}

func (p ProjectApiController) UploadProjectFile(c echo.Context, form *multipart.Form, project *domain.Project) (err error) {
	for name, value := range form.File {
		if !(name == "main_image_file" || name == "contract_before_page" || name == "contract_after_page" || name == "silent_partnerships" || name == "project_report" ||
			strings.Contains(name, "sub_image_files")) {
			if value != nil {
				project.Attachments = make([]*domain.Attachment, 0)
				for _, mFile := range value {
					attachment := &domain.Attachment{
						FileName:      s3utils.MakeS3FileName(),
						OriginalName:  mFile.Filename,
						ResourceId:    project.CommonId,
						FileSize:      mFile.Size,
						FileExtension: s3utils.GetExtension(mFile.Filename),
						FilePath:      s3utils.MakeS3Path(domain.ATTACHMENT_FILE_TYPE), // TYPE 에 따라서 PATH 지정됨
						FileUrl:       "",                                              // CDN 주소와 결합하여 URL 등록
					}
					attachment.FileUrl, err = s3utils.UploadAttachmentFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, attachment, mFile)
					if err != nil {
						return factory.ErrorDB.New(err)
					}
					project.Attachments = append(project.Attachments, attachment)

					if _, err := attachment.Create(c.Request().Context()); err != nil {
						return factory.ErrorDB.New(err)
					}
				}
			}
		} else if name == "contract_before_page" || name == "contract_after_page" || name == "silent_partnerships" || name == "project_report" {
			fileType := domain.CONTRACT_BEFORE_PAGE
			if name == "contract_after_page" {
				fileType = domain.CONTRACT_AFTER_PAGE
				project.ContractAfterPageDate = jsonutils.JSONTime(time.Now())
			} else if name == "silent_partnerships" {
				fileType = domain.SILENT_PARTNERSHIPS
				project.SilentPartnershipsDate = jsonutils.JSONTime(time.Now())
			} else if name == "project_report" {
				fileType = domain.PROJECT_REPORT
				project.ProjectReportDate = jsonutils.JSONTime(time.Now())
			} else if name == "contract_before_page" {
				project.ContractBeforePageDate = jsonutils.JSONTime(time.Now())
			}
			for _, mFile := range value {
				projectFile := &domain.ProjectFile{
					FileName:      s3utils.MakeS3FileName(),
					FileType:      fileType,
					OriginalName:  mFile.Filename,
					ProjectId:     project.Id,
					OwnerId:       project.Id,
					UpdateId:      project.Id,
					FileSize:      mFile.Size,
					FileExtension: s3utils.GetExtension(mFile.Filename),
					FilePath:      s3utils.MakeS3Path(domain.PROJECT_FILE),
					FileUrl:       "",
				}
				_, err = s3utils.UploadAttachmentFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, projectFile, mFile)
				if err != nil {
					return factory.ErrorAWSUpload.New(err)
				}

				if _, err := projectFile.Create(c.Request().Context()); err != nil {
					return factory.ErrorDB.New(err)
				}
			}
		}
	}
	return nil
}

func (p ProjectApiController) createProjectNews(c echo.Context, newsInputs *[]ProjectNewsInput, beforeNews []*domain.ProjectNews, receivers []string, project *domain.Project, actorId int64) error {

	for _, value := range *newsInputs {
		value.ProjectId = project.Id
		if err := c.Validate(&value); err != nil {
			return factory.ErrorDB.New(err, "failed to validate project news")
		}

		projectNews, err := value.ToModel()
		if err != nil {
			return factory.ErrorDB.New(err, "failed to convert project news")
		}

		projectNews.OwnerId = actorId
		projectNews.UpdateId = actorId
		if _, err := projectNews.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err, "failed to create project news")
		}

		if beforeNews != nil {
			// MIR-1492: メールは取り消せないため、保存が確定してから送る。保存が
			// 途中のエラー(同時編集の検知など)で巻き戻った場合は送らず、
			// 送り直した保存で二重に送られることもない。
			// 確定後の処理はリクエストの ctx を引き継ぐため、確定直後に画面が切断されても
			// 送信が打ち切られないよう、キャンセルを外して送る(保存し直しても再送されない)。
			if err := factory.AfterCommit(c.Request().Context(), func(afterCommitCtx context.Context) error {
				p.sendProjectNewsMail(context.WithoutCancel(afterCommitCtx), receivers, projectNews, beforeNews, project)
				return nil
			}); err != nil {
				return factory.ErrorSystem.New(err)
			}
		}
	}

	return nil
}

func (p ProjectApiController) sendProjectNewsMail(ctx context.Context, receivers []string, projectNews *domain.ProjectNews, beforeNews []*domain.ProjectNews, project *domain.Project) {
	// 자동메일
	// TODO 코드 리팩토링 필요
	if len(receivers) > 0 && projectNews.Posted {
		autoMail := true
		for _, news := range beforeNews {
			if projectNews.Title == news.Title && news.Posted {
				autoMail = false
				break
			}
		}

		if autoMail {
			// 발송
			var mail domain.Mail
			err := mail.ProjectNewsMail(ctx, serviceutils.GetServiceInfoMail(), project.Name, projectNews.Title, projectNews.ContentHtml, projectNews.PostDate.Time().Format(time.DateOnly), receivers, project.Id)
			if err != nil {
				factory.Logger(ctx).WithFields(logrus.Fields{
					"err": err,
				}).Warning("mail sending fail")
			}
		}
	}
}

// deleteS3FileAfterCommit は、S3 の実ファイルの削除を保存の確定後に回す(MIR-1492)。
// 保存が途中のエラー(同時編集の検知など)で巻き戻ると、DB は削除前のファイルを
// 参照したままになるため、先に消すとファンドページの画像・書面が表示されなくなる。
// 削除に失敗した場合は従来どおりログを残すだけにする(参照されないファイルが
// S3 に残るだけで、表示には影響しない)。リクエストの ctx を引き継ぐため、確定直後に
// 画面が切断されても打ち切られないよう、キャンセルを外して削除する。
func deleteS3FileAfterCommit(ctx context.Context, file s3utils.S3Object, logFields logrus.Fields, logMessage string) error {
	return factory.AfterCommit(ctx, func(afterCommitCtx context.Context) error {
		deleteCtx := context.WithoutCancel(afterCommitCtx)
		if err := s3utils.DeleteFile(deleteCtx, s3utils.BUCKET_PUBLIC_NAME, file); err != nil {
			logFields["err"] = err
			factory.Logger(deleteCtx).WithFields(logFields).Error(logMessage)
		}
		return nil
	})
}

func (p ProjectApiController) createNewsInputs(c echo.Context, form *multipart.Form, project *domain.Project) (*[]ProjectNewsInput, error) {
	newsInputs := make([]ProjectNewsInput, 0)
	for key, values := range form.Value {
		if !strings.Contains(key, "project_news") {
			continue
		}

		ns := strings.NewReplacer("project_news[", "", "]", "")
		ss := strings.Split(ns.Replace(key), ".")

		index, err := strconv.Atoi(ss[0])
		if err != nil {
			continue
		}
		field := ss[1]
		value := values[0]

		for index >= len(newsInputs) {
			newsInputs = append(newsInputs, ProjectNewsInput{})
		}

		switch field {
		case "title":
			newsInputs[index].Title = value
		case "content_html":
			value, err = p.contentImageService.ReplaceContentImage(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, value, project.CommonId, formutils.FromToUrls(*form, "content_images"))
			if err != nil {
				return nil, factory.ErrorDB.New(err, "failed to replace project news content image")
			}
			newsInputs[index].ContentHtml = value
		case "posted":
			newsInputs[index].Posted, _ = strconv.ParseBool(value)
		case "post_date":
			newsInputs[index].PostDate = value
		}
	}

	return &newsInputs, nil
}

func (ProjectApiController) GetOne(c echo.Context) error {
	var err error
	_, _, err = jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	v, err := domain.Project{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if v == nil {
		return factory.ErrorNotFound.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, ProjectAdminResponseDto{}.New(v))
}
func (p ProjectApiController) projectContentHtmlProcess(c echo.Context, project *domain.Project, form *multipart.Form) error {
	var err error
	// Process Content Images
	project.CompanyInfoHtml, err = p.contentImageService.ReplaceContentImage(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, project.CompanyInfoHtml, project.CommonId, formutils.FromToUrls(*form, "content_images"))
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	project.ProjectInfoHtml, err = p.contentImageService.ReplaceContentImage(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, project.ProjectInfoHtml, project.CommonId, formutils.FromToUrls(*form, "content_images"))
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	project.CautionInfoHtml, err = p.contentImageService.ReplaceContentImage(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, project.CautionInfoHtml, project.CommonId, formutils.FromToUrls(*form, "content_images"))
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	return nil
}

func (p ProjectApiController) mainImageUpload(c echo.Context, updateId int64, imgFile *multipart.FileHeader, project *domain.Project) error {
	if imgFile != nil {
		projectImages, err := domain.ProjectImage{}.FindByProjectIdAndImageType(c.Request().Context(), project.Id, domain.MAIN_IMG)
		if err != nil {
			return factory.ErrorDB.New(err, "failed get project main images")
		}
		if len(projectImages) > 0 {
			for _, img := range projectImages {
				if err := deleteS3FileAfterCommit(c.Request().Context(), img, logrus.Fields{
					"user":       updateId,
					"project_id": project.Id,
					"image_id":   img.Id,
					"image_type": img.ImageType,
				}, "Fail S3 Delete project sub image"); err != nil {
					return factory.ErrorSystem.New(err)
				}
				if err := (domain.ProjectImage{}).Delete(c.Request().Context(), img.Id); err != nil {
					return factory.ErrorDB.New(err, "failed delete project sub image")
				}
			}
		}

		project.MainImage = &domain.ProjectImage{
			FileName:      s3utils.MakeS3FileName(),
			OriginalName:  imgFile.Filename,
			ProjectId:     project.Id,
			FileSize:      imgFile.Size,
			UpdateId:      updateId,
			ImageType:     domain.MAIN_IMG,
			FileExtension: s3utils.GetExtension(imgFile.Filename),
			FilePath:      s3utils.MakeS3Path(domain.PROJECT_IMG_FILE_TYPE), // Type 에 따라서 PATH 지정됨
			FileUrl:       "",                                               // CDN 주소와 결합하여 URL 등록
		}
		project.MainImageUrl, err = s3utils.UploadAttachmentFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, project.MainImage, imgFile)
		if err != nil {
			return factory.ErrorAWSUpload.New(err)
		}

		if _, err := project.MainImage.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err)
		}
		project.MainImageId = project.MainImage.Id
	}
	return nil
}

func (p ProjectApiController) uploadSubImage(
	c echo.Context, updateId int64, imgFiles map[string]*multipart.FileHeader, project *domain.Project, subImageOrder map[string]int64,
) error {
	if len(imgFiles) == 0 {
		return nil
	}

	insertList := []*domain.ProjectImage{}
	for i, imgFile := range imgFiles {
		projectImage := &domain.ProjectImage{
			FileName:      s3utils.MakeS3FileName(),
			OriginalName:  imgFile.Filename,
			ProjectId:     project.Id,
			FileSize:      imgFile.Size,
			UpdateId:      updateId,
			ImageType:     domain.SUB_IMG,
			FileExtension: s3utils.GetExtension(imgFile.Filename),
			FilePath:      s3utils.MakeS3Path(domain.PROJECT_IMG_FILE_TYPE),
			FileUrl:       "",
			Order:         subImageOrder[i],
		}

		url, err := s3utils.UploadAttachmentFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, projectImage, imgFile)
		if err != nil {
			return factory.ErrorAWSUpload.New(err)
		}
		projectImage.FileUrl = url
		insertList = append(insertList, projectImage)
	}

	_, err := domain.ProjectImage{}.CreateList(c.Request().Context(), insertList)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

// shouldAttemptReopen は、終了日時が変更され、かつ現時点より後になったかを
// 判定する(MIR-1314)。前倒しでも変更後の日付が未来ならまだ募集期間中なので
// 対象になる。終了日を変えずに保存しただけを対象に含めると、早期募集終了した
// ファンドが保存のたびに再開されてしまうため、変更の有無で判断する。
//
// beforeEndAt は DB 由来、afterEndAt はフォーム由来で生成元が異なる。Equal は
// 同じ瞬間かを見るため場所の違いは吸収されるが、保存形式やタイムゾーンの扱いを
// 変える際は、無変更を変更と誤判定しないか確かめること。
func shouldAttemptReopen(beforeEndAt, afterEndAt, now time.Time) bool {
	if afterEndAt.Equal(beforeEndAt) {
		return false
	}
	return afterEndAt.After(now)
}

// shouldReopenFunding は、画面(status)が募集終了を表示していた状態での延長かを
// 判定する(MIR-1314)。shouldAttemptReopenと違い画面の状態も見るため、
// 「延長したのに再開できなかった」場合にエラーとして扱うべきかの判定にしか
// 使わない。募集を試みるかどうかの判定自体はshouldAttemptReopenだけで行う
// (画面が古い状態(募集中等)を表示していても、DBが既に募集終了になっていれば
// 再開を試みる必要があるため)。
func shouldReopenFunding(
	status domain.PROJECT_STATUS,
	beforeEndAt, afterEndAt, now time.Time,
) bool {
	return status == domain.END_FUNDING && shouldAttemptReopen(beforeEndAt, afterEndAt, now)
}

// projectReopenAlreadyConflicted は、画面が募集終了を表示しており延長は
// 再開のつもりだったのに、保存の直前に読み直したDBの状態が既に募集終了で
// なくなっているかを判定する(MIR-1314)。trueの場合、呼び出し元はS3書き込み・
// メール送信などの取り消せない副作用が始まる前に409を返す。
func projectReopenAlreadyConflicted(explicitReopenIntent bool, currentStatus domain.PROJECT_STATUS) bool {
	return explicitReopenIntent && currentStatus != domain.END_FUNDING
}

// projectReopenResult は、ReopenFundingの実行結果(reopened)を、保存が成功と
// 言えるか(conflict=false)、409にすべきか(conflict=true)に変換する
// (MIR-1314)。画面が募集終了を表示しておらず再開のつもりが無かった場合は、
// 再開できなくても保存は成功として扱う。
func projectReopenResult(reopened, explicitReopenIntent bool) (conflict bool) {
	return !reopened && explicitReopenIntent
}

// canDeleteProjectFile は、fileType の書面を削除してよいかを判定する。
//
// 必須書面(RequiredProjectFileTypes)は空にできない仕様のため、同じ
// リクエストで同じ種別の新規ファイルが送られている(差し替え)場合のみ
// 削除を許可する。必須でない種別(ファンドレポート等)は常に許可する。
func canDeleteProjectFile(fileType domain.PROJECT_FILE_TYPE, newFiles map[string][]*multipart.FileHeader) bool {
	formKey, required := domain.RequiredProjectFileTypes[fileType]
	if !required {
		return true
	}
	return len(newFiles[formKey]) > 0
}

// validateProjectFileDeleteRequest は project_file_delete で指定された
// 削除対象を、実際に何かを変更する前に検証する。
//
// このあと Update はサブ画像の S3 実削除・既存ニュースの削除・投資家への
// メール送信など、DB トランザクションのロールバックでは取り消せない
// 副作用を実行する。削除対象の検証をそれより後段(実削除の直前)で行うと、
// 必須書面を差し替えなしで消そうとしただけのリクエストがエラーになった
// 場合に、これらの副作用だけが実行済みのまま残ってしまう。
func validateProjectFileDeleteRequest(ctx context.Context, form *multipart.Form, projectId int64) error {
	for key, values := range form.Value {
		if !strings.Contains(key, "project_file_delete") {
			continue
		}
		for _, value := range values {
			fileId, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return factory.ErrorParameter.New(err)
			}
			projectFile, err := domain.ProjectFile{}.GetById(ctx, fileId)
			if err != nil {
				return factory.ErrorDB.New(err, "failed to get project file")
			}
			if projectFile == nil || projectFile.ProjectId != projectId {
				// 対象が見つからない/他プロジェクトの場合は、実削除ループ側の
				// 既存ログでエラーとして記録される。ここでは弾く必要が無い。
				continue
			}
			if !canDeleteProjectFile(projectFile.FileType, form.File) {
				return factory.ErrorParameter.New(errors.New(
					"required project file cannot be deleted without a replacement: " + string(projectFile.FileType)))
			}
		}
	}
	return nil
}

func (p ProjectApiController) Update(c echo.Context) error {
	var v ProjectInput
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	uid, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if err := c.Validate(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	project, err := v.ToModel()
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	project.Id = id
	project.UpdateId = uid

	beforeProject, err := domain.Project{}.GetById(c.Request().Context(), project.Id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if beforeProject == nil {
		return factory.ErrorNotFound.New(nil)
	}

	// MIR-1492: 画面を開いた後に別の保存があれば、このまま保存すると相手の変更を
	// 巻き戻すため、S3書き込み・メール送信などの副作用より前に409で知らせ、この
	// 保存は止める。上書きを禁止はしない(画面は確認のうえ force_overwrite を付けて
	// 送り直せる)。ここはロックしない読み取りのため、ほぼ同時の保存は書き込みの
	// 時点(UpdateForAdminDetail)でも確かめる。
	expectedDetailVersion, err := ExpectedDetailVersion(v.DetailVersion, v.ForceOverwrite)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	if IsDetailVersionConflicted(expectedDetailVersion, beforeProject.DetailVersion) {
		return factory.ErrorProjectDetailUpdatedAfterLoad.New(errors.New("project detail already updated after load"))
	}

	project.CommonId = beforeProject.CommonId
	project.IsWaitInvest = beforeProject.IsWaitInvest

	// MIR-1314: 終了日時の延長による募集再開の判定は、副作用が何も始まる前の
	// ここで済ませる。attemptReopenは延長したかどうかだけを見て「再開を
	// 試みるか」を、explicitReopenIntentは画面が募集終了を表示していたかを
	// 見て「再開できなかった場合にエラーにするか」を判定する
	// (shouldAttemptReopen/shouldReopenFundingのコメント参照)。画面が募集
	// 終了を表示しているのに、ここで読み直したDBが既に募集終了でなくなって
	// いれば、この時点で競合と確定するため、S3書き込み・メール送信などの
	// 取り消せない副作用より前に409で弾く。
	now := time.Now()
	attemptReopen := shouldAttemptReopen(beforeProject.EndAt.Time(), project.EndAt.Time(), now)
	explicitReopenIntent := shouldReopenFunding(project.ProjectStatus, beforeProject.EndAt.Time(), project.EndAt.Time(), now)
	if projectReopenAlreadyConflicted(explicitReopenIntent, beforeProject.ProjectStatus) {
		return factory.ErrorProjectStatusChanged.New(errors.New("project status already changed"))
	}

	// project_file_delete の検証は、S3 への書き込み・実ファイル削除や
	// ニュースメール送信など元に戻せない副作用が何ひとつ始まる前に行う
	// (直後の projectContentHtmlProcess も本文中の画像を S3 へコピーする)。
	// DB はミドルウェアのトランザクションでロールバックされるが、これらの
	// 外部副作用はロールバックされないため、途中で弾くと不整合な状態が
	// 残ってしまう。
	if err := validateProjectFileDeleteRequest(c.Request().Context(), form, project.Id); err != nil {
		return err
	}
	// MIR-1403: attachments_delete も、S3を触る処理より前に、数字でない値が無いかを確かめておく。
	if _, err := domain.ParseAttachmentDeleteIds(form.Value); err != nil {
		return factory.ErrorParameter.New(err)
	}

	if err := p.projectContentHtmlProcess(c, project, form); err != nil {
		return factory.ErrorDB.New(errors.New("failed project content html process"))
	}

	exist, err := project.CheckExistClosedSlug(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if exist {
		return factory.ErrorAlreadyExists.New(errors.New("project with the same closed slug already exists"))
	}

	existFundNumber, err := project.CheckExistFundNumber(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if existFundNumber {
		return factory.ErrorAlreadyExists.New(errors.New("project with the same fund number already exists"))
	}

	_, mImgFile, _ := c.Request().FormFile("main_image_file")
	if err := p.mainImageUpload(c, uid, mImgFile, project); err != nil {
		return factory.ErrorAWSUpload.New(errors.New("failed project main image uploads"))
	}

	subImages := map[string]*multipart.FileHeader{}
	for name, value := range form.File {
		if strings.Contains(name, "sub_image_files") && len(value) > 0 {
			if index, ok := extractSubImageIndex(name); ok {
				subImages[index] = value[0]
			}
		}
	}

	subImageOrder := map[string]int64{}
	subImageIDs := map[int64]string{}
	for name, value := range form.Value {
		if strings.Contains(name, "sub_image_order") && len(value) > 0 {
			if index, ok := extractSubImageIndex(name); ok {
				order, err := strconv.Atoi(value[0])
				if err != nil {
					return factory.ErrorParameter.New(err)
				}
				subImageOrder[index] = int64(order)
			}
		} else if strings.Contains(name, "sub_image_files") && len(value) > 0 {
			if strings.Contains(name, "].id") {
				index, ok := extractSubImageIndex(name)
				if !ok {
					continue
				}
				subImageID, err := strconv.ParseInt(value[0], 10, 64)
				if err != nil {
					return factory.ErrorParameter.New(err)
				}
				subImageIDs[subImageID] = index
			}
		}
	}

	projectImages, err := domain.ProjectImage{}.FindByProjectId(c.Request().Context(), project.Id)
	if err != nil {
		return factory.ErrorDB.New(err, "failed get project sub images")
	}
	for _, img := range projectImages {
		if img.ImageType != domain.SUB_IMG {
			continue
		}
		if _, ok := subImageIDs[img.Id]; ok {
			img.Order = subImageOrder[subImageIDs[img.Id]]
			if _, err := img.Update(c.Request().Context()); err != nil {
				return factory.ErrorDB.New(err, "failed delete project sub image")
			}
			continue
		}

		if err := deleteS3FileAfterCommit(c.Request().Context(), img, logrus.Fields{
			"user":       uid,
			"project_id": project.Id,
			"image_id":   img.Id,
			"image_type": img.ImageType,
		}, "Fail S3 Delete project sub image"); err != nil {
			return factory.ErrorSystem.New(err)
		}

		if err := (domain.ProjectImage{}).Delete(c.Request().Context(), img.Id); err != nil {
			return factory.ErrorDB.New(err, "failed delete project sub image")
		}
	}

	if len(subImages) > 0 {
		if err := p.uploadSubImage(c, uid, subImages, project, subImageOrder); err != nil {
			return factory.ErrorAWSUpload.New(errors.New("failed project sub image uploads"))
		}
	}

	updatedProjectImages, err := domain.ProjectImage{}.FindByProjectIdAndImageType(c.Request().Context(), project.Id, domain.SUB_IMG)
	if err != nil {
		return factory.ErrorDB.New(err, "failed get project main images")
	}
	for _, img := range updatedProjectImages {
		project.SubImageInfoList = append(project.SubImageInfoList, &domain.SubImageInfo{
			SubImageId:  img.Id,
			SubImageUrl: img.FileUrl,
			SubImage:    img,
		})
	}

	_, newsList, err := domain.ProjectNews{}.GetProjectById(c.Request().Context(), project.Id, "", "", 0, 30)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	for _, news := range newsList {
		err := domain.ProjectNews{}.Delete(c.Request().Context(), news.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}
	attachmentFilesToDelete, err := deleteOwnedAttachments(c.Request().Context(), form.Value, project.CommonId,
		logrus.Fields{"user": uid, "project_id": project.Id})
	if err != nil {
		return err
	}

	receivers, err := p.GetProjectInvestorMails(c, project)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	newsInputs, err := p.createNewsInputs(c, form, project)
	if err != nil {
		return err
	}

	err = p.createProjectNews(c, newsInputs, newsList, receivers, project, uid)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	milestoneInputs, err := p.createMilestoneInputs(form)
	if err != nil {
		return err
	}

	projectMilestones, err := domain.ProjectMilestone{}.GetAllByProjectId(c.Request().Context(), project.Id)
	if err != nil {
		return factory.ErrorDB.New(err, "failed get project milestones")
	}
	for _, milestone := range projectMilestones {
		err := domain.ProjectMilestone{}.Delete(c.Request().Context(), milestone.Id)
		if err != nil {
			return factory.ErrorDB.New(err)
		}
	}

	err = p.createMilestone(c, milestoneInputs, project)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	for key, values := range form.Value {
		if strings.Contains(key, "project_file_delete") {
			for _, value := range values {
				fileId, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return factory.ErrorDB.New(err)
				}
				projectFile, err := domain.ProjectFile{}.GetById(c.Request().Context(), fileId)
				if err != nil {
					return factory.ErrorDB.New(err, "failed to get project file")
				}
				if projectFile == nil || projectFile.ProjectId != project.Id {
					factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
						"user":      uid,
						"id":        fileId,
						"file_type": "project_file",
					}).Error("Attempted to delete project file not belonging to this project")
					continue
				}
				// 必須書面の削除可否は validateProjectFileDeleteRequest で
				// 副作用が始まる前に検証済み。ここでは実削除のみ行う。
				err = domain.ProjectFile{}.Delete(c.Request().Context(), fileId)
				if err != nil {
					factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
						"user":      uid,
						"id":        fileId,
						"err":       err,
						"file_type": "project_file",
					}).Error("Fail S3 Delete")
				}

			}
		}
	}

	for name, value := range form.File {
		if !(name == "main_image_file" || name == "contract_before_page" || name == "contract_after_page" || name == "silent_partnerships" || name == "project_report" ||
			strings.Contains(name, "sub_image_files")) {
			if value != nil {
				project.Attachments = make([]*domain.Attachment, 0)
				for _, mFile := range value {
					attachment := &domain.Attachment{
						FileName:      s3utils.MakeS3FileName(),
						OriginalName:  mFile.Filename,
						ResourceId:    project.CommonId,
						FileSize:      mFile.Size,
						FileExtension: s3utils.GetExtension(mFile.Filename),
						FilePath:      s3utils.MakeS3Path(domain.ATTACHMENT_FILE_TYPE), // Type 에 따라서 PAHT 지정됨
						FileUrl:       "",                                              // CDN 주소와 결합하여 URL 등록
					}
					attachment.FileUrl, err = s3utils.UploadAttachmentFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, attachment, mFile)
					if err != nil {
						return factory.ErrorAWSUpload.New(err)
					}
					project.Attachments = append(project.Attachments, attachment)

					if _, err := attachment.Create(c.Request().Context()); err != nil {
						return factory.ErrorDB.New(err)
					}
				}
			}
		} else if name == "contract_before_page" || name == "contract_after_page" || name == "silent_partnerships" || name == "project_report" {
			fileType := domain.CONTRACT_BEFORE_PAGE
			if name == "contract_after_page" {
				fileType = domain.CONTRACT_AFTER_PAGE
				project.ContractAfterPageDate = jsonutils.JSONTime(time.Now())
			} else if name == "silent_partnerships" {
				fileType = domain.SILENT_PARTNERSHIPS
				project.SilentPartnershipsDate = jsonutils.JSONTime(time.Now())
			} else if name == "project_report" {
				fileType = domain.PROJECT_REPORT
				project.ProjectReportDate = jsonutils.JSONTime(time.Now())
			} else if name == "contract_before_page" {
				project.ContractBeforePageDate = jsonutils.JSONTime(time.Now())
			}
			for _, mFile := range value {
				projectFile := &domain.ProjectFile{
					FileName:      s3utils.MakeS3FileName(),
					FileType:      fileType,
					OriginalName:  mFile.Filename,
					ProjectId:     project.Id,
					OwnerId:       project.Id,
					UpdateId:      project.Id,
					FileSize:      mFile.Size,
					FileExtension: s3utils.GetExtension(mFile.Filename),
					FilePath:      s3utils.MakeS3Path(domain.PROJECT_FILE),
					FileUrl:       "",
				}
				_, err = s3utils.UploadAttachmentFile(c.Request().Context(), s3utils.BUCKET_PUBLIC_NAME, projectFile, mFile)
				if err != nil {
					return factory.ErrorAWSUpload.New(err)
				}

				if _, err := projectFile.Create(c.Request().Context()); err != nil {
					return factory.ErrorDB.New(err)
				}
			}
		}
	}

	// MIR-1314: 状態(project_status)・キャンセル待ちフラグ(is_wait_invest)は
	// このハンドラでは書かない。画面にこの2項目を編集する欄は無く、開いた
	// 時点の値をフォームの初期値としてそのまま送り返しているだけなので、
	// 書き戻すと並行して動くジョブ・状態変更ボタンによる変更を古い値へ
	// 戻してしまう。終了日時の延長による募集再開だけは、保存後にReopenFunding
	// で別途行う。
	updated, err := project.UpdateForAdminDetail(c.Request().Context(), expectedDetailVersion)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if expectedDetailVersion != nil && !updated {
		// MIR-1492: 事前チェックのあと、この処理の途中(S3書き込み等を含む)で別の
		// 保存が先に確定した。エラーを返してトランザクションごと巻き戻す
		// (お知らせのメールは確定後に送るため、ここでは送られていない)。
		return factory.ErrorProjectDetailUpdatedAfterLoad.New(errors.New("project detail updated by a concurrent save"))
	}

	reopened := false
	if attemptReopen {
		ok, err := project.ReopenFunding(c.Request().Context())
		if err != nil {
			return factory.ErrorDB.New(err)
		}
		reopened = ok
		if projectReopenResult(reopened, explicitReopenIntent) {
			// 事前チェックのあと、この処理の途中(S3書き込み等を含む)で
			// 別のリクエストやジョブがさらに状態を変えた。最後の砦として弾く。
			return factory.ErrorProjectStatusChanged.New(errors.New("project status already changed"))
		}
	}

	afterProject, err := domain.Project{}.GetById(c.Request().Context(), project.Id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	// レスポンス・変更履歴用に、実際にDBへ反映された状態を控えておく
	// (次の行以降で履歴の比較用にafterProject.ProjectStatusを保存前の値へ
	// 正規化するため、その前の実際の値をここで退避する)。
	actualStatus := afterProject.ProjectStatus

	// 保存後のメールは送らない(このハンドラで変わりうる状態は募集再開だけで、
	// 募集中への遷移はメール対象外)。並行して動くジョブ・状態変更ボタンに
	// よる状態の変化を、この保存操作による変更として履歴に誤って残さない
	// よう、再開でなければ読み直した状態を保存前の値へそろえてから比較する。
	if !reopened {
		afterProject.ProjectStatus = beforeProject.ProjectStatus
	}
	project.ProjectStatus = actualStatus
	// MIR-1492: 画面が次の保存で送り返す値として、保存後の保存回数を返す。
	project.DetailVersion = afterProject.DetailVersion

	_, err = domain.ProjectInfoHistory{}.Create(c.Request().Context(), beforeProject, afterProject, uid)
	if err != nil {
		return err
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), uid, email,
		"menu.manage_product.product_detail", project.Name, c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	domain.DeleteAttachmentFilesAfterCommit(c.Request().Context(), attachmentFilesToDelete)

	return apirender.RenderSucc(c, http.StatusOK, ProjectAdminResponseDto{}.New(project))
}

func (p ProjectApiController) GetProjectInvestorMails(c echo.Context, project *domain.Project) ([]string, error) {
	// 투자대기중 이거나 투자완료한 투자자 목록
	investors, err := project.GetAcceptedNewsletterAndCompletedInvestors(c.Request().Context())
	if err != nil {
		return nil, factory.ErrorDB.New(err)
	}

	receivers := make([]string, 0, 0)
	for _, user := range investors {
		if user.Email != "" {
			receivers = append(receivers, user.Email)
		}
	}
	return receivers, nil
}

func (ProjectApiController) Delete(c echo.Context) error {
	var err error
	_, _, err = jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = domain.Project{}.Delete(c.Request().Context(), id)

	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, nil)
}

func (ProjectApiController) Download(c echo.Context) error {
	var v ProjectSearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	userId, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return factory.ErrorTokenInvalid.New(err)
	}

	// console log
	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	err = domain.Project{}.DownloadCsv(c.Request().Context(), c.Response().Writer,
		v.Sortby, v.Order, v.ProjectId, v.ProjectStatus, v.TargetShareAmount, v.ProjectName, v.ProjectType, v.CategoryType, v.BlockInvest, v.StartDate, v.EndDate, v.OutOfDateProject, v.Posted)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), userId, email,
		"menu.manage_product.product_list", "", c.RealIP(), domain.DOWNLOAD)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return nil
}

// CheckDuplicateFundNumber は入力中のファンド番号が使用済みかを返す。
// 保存時にも同じ判定を行うが、入力の時点で気付けるように画面から呼ばれる。
func (ProjectApiController) CheckDuplicateFundNumber(c echo.Context) error {
	var v CheckDuplicateFundNumberInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}

	fundNumber, err := ParseFundNumber(v.FundNumber)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	project := domain.Project{Id: v.ProjectId, FundNumber: fundNumber}

	isDuplicated, err := project.CheckExistFundNumber(c.Request().Context())
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, CheckDuplicateFundNumberResponseDto{IsDuplicated: isDuplicated})
}

func (ProjectApiController) UpdateProjectStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}
	projectStatus := c.Param("project_status")

	uid, email, err := jwthelper.TokenToUser(c)
	if err != nil {
		return err
	}

	beforeProject, err := domain.Project{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	if beforeProject == nil {
		return factory.ErrorNotFound.New(err)
	}

	project := *beforeProject

	ctx := c.Request().Context()
	// 償還完了で「元本が残っている・償還予定が無い」確認に画面が同意したときだけ、進めてよい(MIR-272)。
	if c.QueryParam("confirm_remaining_principal") == "true" {
		ctx = domain.WithRemainingPrincipalConfirmed(ctx)
	}

	if err := project.ChangeStatus(ctx, domain.PROJECT_STATUS(projectStatus), uid); err != nil {
		return errors.Wrap(err, "fail to change project status")
	}

	err = domain.OperatorHistory{}.Record(c.Request().Context(), uid, email,
		"menu.manage_product.product_detail", project.Name, c.RealIP(), domain.CHANGE)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, project)
}

// GetCloseCheck は、償還完了にする前の確認の材料(未実施の予定の件数・最後の予定の元本残高)を返す
// (MIR-272)。画面が確認の要否(confirmation_required)を決め、確認の文言に残元本を出すために使う。状態は変えない。
func (ProjectApiController) GetCloseCheck(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	project, err := domain.Project{}.GetById(c.Request().Context(), id)
	if err != nil {
		return factory.ErrorDB.New(err)
	}
	if project == nil {
		return factory.ErrorNotFound.New(errors.New("project not found"))
	}

	closeCheck, err := domain.Project{}.GetCloseCheck(c.Request().Context(), id, false)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, closeCheck)
}

func (ProjectApiController) GetInvestmentsByProjectId(c echo.Context) error {
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	projectId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	totalCount, invests, err := domain.Investment{}.
		GetInvestmentsByProjectId(c.Request().Context(), projectId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorDB.New(err)
	}

	return apirender.RenderSucc(c, http.StatusOK, apirender.ArrayResult{
		Items:      invests,
		TotalCount: totalCount,
	})
}

func (ProjectApiController) InvestmentsDownload(c echo.Context) error {
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return factory.ErrorParameter.New(err)
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}

	factory.Logger(c.Request().Context()).WithFields(logrus.Fields{
		"sortby":           v.Sortby,
		"order":            v.Order,
		"max_result_count": v.MaxResultCount,
		"skip_count":       v.SkipCount,
	}).Info("SearchStart")

	projectId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return factory.ErrorParameter.New(err)
	}

	err = domain.Investment{}.InvestmentsDownloadCsv(c.Request().Context(), c.Response().Writer, projectId, v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return factory.ErrorFailDownloadCSV.New(err)
	}

	return nil
}

func (ProjectApiController) createMilestoneInputs(form *multipart.Form) (*[]ProjectMilestoneInput, error) {
	var milestoneInputs []ProjectMilestoneInput
	for key, values := range form.Value {
		if !strings.Contains(key, "project_milestones") {
			continue
		}

		ns := strings.NewReplacer("project_milestones[", "", "]", "")
		ss := strings.Split(ns.Replace(key), ".")

		index, err := strconv.Atoi(ss[0])
		if err != nil {
			continue
		}
		field := ss[1]
		value := values[0]

		for index >= len(milestoneInputs) {
			milestoneInputs = append(milestoneInputs, ProjectMilestoneInput{})
		}

		switch field {
		case "milestone_no":
			milestoneInputs[index].MilestoneNo, err = strconv.Atoi(value)
			if err != nil {
				return nil, factory.ErrorDB.New(err, "fail to parse milestone_no")
			}
		case "milestone_info":
			milestoneInputs[index].MilestoneInfo = value
		case "milestone_assignment":
			milestoneInputs[index].MilestoneAssignment, err = strconv.Atoi(value)
			if err != nil {
				return nil, factory.ErrorDB.New(err, "fail to parse milestone assignment")
			}
		case "milestone_status":
			milestoneInputs[index].MilestoneStatus = value
		case "milestone_note":
			milestoneInputs[index].MilestoneNote = value
		}
	}

	return &milestoneInputs, nil
}

func (p ProjectApiController) createMilestone(c echo.Context, inputs *[]ProjectMilestoneInput, project *domain.Project) error {
	for _, value := range *inputs {
		if err := c.Validate(&value); err != nil {
			return factory.ErrorDB.New(err, "fail to validate milestone")
		}
		item, err := value.ToModel()
		if err != nil {
			return factory.ErrorDB.New(err, "failed to convert milestone to model")
		}
		item.ProjectId = project.Id

		if _, err := item.Create(c.Request().Context()); err != nil {
			return factory.ErrorDB.New(err, "failed to create milestone")
		}
	}
	return nil
}

func extractSubImageIndex(name string) (string, bool) {
	start := strings.Index(name, "[")
	if start == -1 {
		return "", false
	}
	end := strings.Index(name[start:], "]")
	if end == -1 {
		return "", false
	}
	if end <= 1 {
		return "", false
	}
	return name[start+1 : start+end], true
}
