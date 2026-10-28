package jobs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/factory"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/jsonutils"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/serviceutils"
	"github.com/illuminarean/energy-shares-backend-repo/service"
)

// Jobでユーザー更新があったことを記録するための定数
const (
	jobProcessId   = 1
	jobProcessName = "User Address Check Job"
)

type UserAddressCheckJobs struct{}

type NexwayError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (UserAddressCheckJobs) UserAddressCheckJob(ctx context.Context) error {
	return runJobWithExecutionHistory(ctx, JobExecutionHistoryInput{
		JobType: domain.JOB_EXECUTION_TYPE_USER_ADDRESS_CHECK,
		JobName: jobNameUserAddressCheck,
	}, userAddressCheckJob)
}

func userAddressCheckJob(ctx context.Context) error {
	userInfoHistoryService := service.NewUserInfoHistoryService()

	logrus.WithFields(logrus.Fields{
		"title": "=======Start=======",
		"desc":  "=======Start Jobs=======",
		"time":  time.Now(),
	}).Info("User Address Check Jobs")

	// 조회되는 유저 -  가입상태 - 주소확인중, 우편물 발송 상태 미 발송
	var corpUserList []domain.CorporateUser

	if err := factory.DB(ctx).Table("user").
		Select("user.* ,"+domain.CORPORATE_USER_COLUMN_STRING).
		Join("INNER", "corporate_profile", "user.id = corporate_profile.user_id").
		Where("user.type = ? AND user.audit_status = ? AND ( user.postal_code_status = ? OR corporate_profile.corporate_postal_status = ? ) ", domain.CORPORATE, domain.CHECK_ADDRESS_AUDIT, false, false).
		Find(&corpUserList); err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "Fail Select Corporate User Table",
			"err":   err,
		}).Error("User Address Check Jobs")
		return err
	}

	if len(corpUserList) == 0 {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Match User Zero=======",
			"time":  time.Now(),
		}).Info("User Address Check Jobs")
		return nil
	}

	// 1. 주문번호 발행
	tempId, err := formRegistration(ctx)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"title":  "=======Success formRegistration=======",
		"desc":   "",
		"time":   time.Now(),
		"tempId": tempId,
	}).Info("User Address Check Jobs")

	// 2. 유저 csv 파일 전달
	_, err = formFileRegistration(ctx, tempId, corpUserList)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"title":  "=======Success formFileRegistration=======",
		"desc":   "",
		"time":   time.Now(),
		"tempId": tempId,
	}).Info("User Address Check Jobs")

	// 3. 주문 실행
	orderId, err := formOrder(ctx, tempId)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"title":   "=======Success formOrder=======",
		"desc":    "",
		"time":    time.Now(),
		"orderId": orderId,
	}).Info("User Address Check Jobs")

	for _, corpUser := range corpUserList {
		before, err := userInfoHistoryService.GetDataForCorporateUser(ctx, corpUser.UserId)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"title": "=======End User Address Check Jobs=======",
				"desc":  "Fail Select Before Corporate User Table",
				"err":   err,
			}).Error("User Address Check Jobs")
			return err
		}

		if !corpUser.PostalCodeStatus {
			corpUser.PostalCodeStatus = true
			corpUser.PostalCodeSendDate = jsonutils.JSONTime(time.Now())
			q := factory.DB(ctx).UseBool("postal_code_status")
			q.ID(corpUser.User.Id).Update(&corpUser.User)
		}

		if !corpUser.CorporatePostalStatus {
			corpUser.CorporateProfile.CorporatePostalStatus = true
			corpUser.CorporatePostalCodeSendDate = jsonutils.JSONTime(time.Now())
			q2 := factory.DB(ctx).UseBool("corporate_postal_status")
			q2.Where("corporate_profile.user_id = ?", corpUser.UserId).Update(&corpUser.CorporateProfile)
		}

		err = (&domain.Mail{}).SignUpAuditAddressCheckForCorporateEmail(ctx, serviceutils.GetServiceInfoMail(), corpUser.Email, corpUser.Name)
		if err != nil {
			factory.Logger(ctx).WithFields(logrus.Fields{
				"auditStatus": corpUser.AuditStatus,
				"err":         err,
				"userId":      corpUser.User.Id,
			}).Error("Failed to send mail (SIGN_UP_AUDIT_ADDRESS_CHECK_FOR_CORPORATE_TEMPLATE)")
			return factory.ErrorAWSSendMail.New(err)
		}

		after, err := userInfoHistoryService.GetDataForCorporateUser(ctx, corpUser.UserId)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"title": "=======End User Address Check Jobs=======",
				"desc":  "Fail Select After Corporate User Table",
				"err":   err,
			}).Error("User Address Check Jobs")
			return err
		}

		userInfoHistoryService.CreateUserInfoHistory(ctx, *before, *after, jobProcessId, jobProcessName, "")
	}

	logrus.WithFields(logrus.Fields{
		"title":   "=======End User Address Check Jobs=======",
		"desc":    "=======Success End Jobs=======",
		"time":    time.Now(),
		"OrderId": orderId,
	}).Info("User Address Check Jobs")

	return nil
}

func formRegistration(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("user", serviceutils.GetServiceNexwayId())
	data.Set("password", serviceutils.GetServiceNexwayPwd())

	req, err := http.NewRequest(http.MethodPost, serviceutils.GetServiceNexwayRegistAPI(), strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Network Error formRegistration=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return "", err
	}

	respBody, err := executeHttp(ctx, req, "formRegistration")
	if err != nil {
		return "", err
	}

	var v struct {
		Result struct {
			ApiVersion string      `json:"apiVersion"`
			TempId     string      `json:"tempId"`
			Error      NexwayError `json:"error"`
		} `json:"result"`
	}

	json.Unmarshal(respBody, &v)
	if v.Result.Error.Code != "" {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Error formRegistration=======",
			"code":  v.Result.Error.Code,
			"Msg":   v.Result.Error.Message,
		}).Error("User Address Check Jobs")
		return "", errors.New("formRegistration Error Code : " + v.Result.Error.Code + ", Msg : " + v.Result.Error.Message)
	}
	return v.Result.TempId, nil
}

func formFileRegistration(ctx context.Context, tempId string, corpList []domain.CorporateUser) (string, error) {
	file, err := os.Create("./user.csv")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======CSV CREATE Error formFileRegistration=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return "", err
	}
	defer file.Close()
	// csv writer 생성

	converter := transform.NewWriter(file, japanese.ShiftJIS.NewEncoder())
	wr := bufio.NewWriter(converter)
	defer wr.Flush()

	for _, user := range corpList {
		if user.PostalCodeStatus == false {
			r := []string{
				user.ZipCode,
				user.Province + user.City + user.Address + *user.Address2,
				user.Name,
				user.Uid,
				"",
				"",
				"",
				"k01",
				"",
				"",
				"",
				"",
				user.Province + user.City,
				user.Address,
				*user.Address2,
				"",
				user.Name + " 様",
				"",
				"",
				"",
				user.PostalCode,
			}
			writeLine(wr, r)
		}
		if user.CorporatePostalStatus == false {
			r := []string{
				user.ManagerZipCode,
				user.ManagerProvince + user.ManagerCity + user.ManagerAddress + *user.ManagerAddress2,
				*user.ManagerFamilyName + *user.ManagerFirstName,
				user.Uid + "MANAGER",
				"",
				"",
				"",
				"k01",
				"",
				"",
				"",
				"",
				user.ManagerProvince + user.ManagerCity,
				user.ManagerAddress,
				*user.ManagerAddress2,
				"",
				user.Name + " 様",
				"",
				"",
				"",
				user.CorporatePostalCode,
			}
			writeLine(wr, r)
		}

	}
	wr.Flush()

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	mw.WriteField("user", serviceutils.GetServiceNexwayId())
	mw.WriteField("password", serviceutils.GetServiceNexwayPwd())
	mw.WriteField("tempId", tempId)
	readFile, err := os.Open("./user.csv")
	if err != nil {
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"title": "=======End User Address Check Jobs=======",
				"desc":  "=======Csv Open Error formFileRegistration=======",
				"err":   err,
			}).Error("User Address Check Jobs")
			return "", err
		}
	}
	defer readFile.Close()
	// -----------------------------------------------
	part, err := mw.CreateFormFile("file", filepath.Base("./user.csv"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Csv Error formFileRegistration=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return "", err
	}
	_, err = io.Copy(part, readFile)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Csv Copy Error formFileRegistration=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return "", err
	}

	mw.Close()
	req, err := http.NewRequest(http.MethodPost, serviceutils.GetServiceNexwayFileAPI(), buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Network Error formFileRegistration=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return "", err
	}

	respBody, err := executeHttp(ctx, req, "formFileRegistration")
	if err != nil {
		return "", err
	}

	var v struct {
		Result struct {
			ApiVersion string      `json:"apiVersion"`
			Status     string      `json:"status"`
			Error      NexwayError `json:"error"`
		} `json:"result"`
	}

	json.Unmarshal(respBody, &v)
	if v.Result.Error.Code != "" {
		logrus.WithFields(logrus.Fields{
			"desc": "=======Error formFileRegistration=======",
			"code": v.Result.Error.Code,
			"Msg":  v.Result.Error.Message,
		}).Error("User Address Check Jobs")
		return "", errors.New("formFileRegistration Error Code : " + v.Result.Error.Code + ", Msg : " + v.Result.Error.Message)
	}
	return "", nil
}

func writeLine(wr *bufio.Writer, r []string) {
	res := ""
	for idx, line := range r {
		res += "\"" + line + "\""
		if idx+1 != len(r) {
			res += ","
		}
	}
	fmt.Fprintln(wr, res)
}

func formOrder(ctx context.Context, tempId string) (string, error) {
	data := url.Values{}
	data.Set("user", serviceutils.GetServiceNexwayId())
	data.Set("password", serviceutils.GetServiceNexwayPwd())
	data.Set("tempId", tempId)

	req, err := http.NewRequest(http.MethodPost, serviceutils.GetServiceNexwayOrderAPI(), strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Network Error formOrder=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return "", err
	}
	respBody, err := executeHttp(ctx, req, "formOrder")
	if err != nil {
		return "", err
	}

	var v struct {
		Result struct {
			ApiVersion string      `json:"apiVersion"`
			OrderNo    string      `json:"orderNo"`
			Error      NexwayError `json:"error"`
		} `json:"result"`
	}

	json.Unmarshal(respBody, &v)
	if v.Result.Error.Code != "" {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Error formOrder=======",
			"code":  v.Result.Error.Code,
			"Msg":   v.Result.Error.Message,
		}).Error("User Address Check Jobs")
		return "", errors.New("formOrder Error Code : " + v.Result.Error.Code + ", Msg : " + v.Result.Error.Message)
	}
	return v.Result.OrderNo, nil
}

func executeHttp(ctx context.Context, req *http.Request, apiName string) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======Network Error " + apiName + "=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"title":         "=======End User Address Check Jobs=======",
			"desc":          "=======Network Error " + apiName + "=======",
			"errStatusCode": resp.StatusCode,
		}).Error("User Address Check Jobs")
		return nil, errors.New(apiName + "Network Error")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"title": "=======End User Address Check Jobs=======",
			"desc":  "=======I/O Error " + apiName + "=======",
			"err":   err,
		}).Error("User Address Check Jobs")
		return nil, err
	}

	return respBody, nil
}
