package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gopkg.in/yaml.v2"

	"github.com/illuminarean/energy-shares-backend-repo/domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/pathutils"
)

// MIR-1408: admin-api の全ルートの権限を表で持ち、本番と同じ組み立て(appOptions・configureEcho)で登録されたルートと
// 突き合わせる。ルートを足した人は、ここに権限を書かないとテストが通らない。

// publicRoute は、トークンの検査を飛ばす(jwt_skip.yml に載せる)ルートの印。
var publicRoute = []domain.ROLE_TYPE{"PUBLIC"}

// routeRoles は「メソッド パス」→ AllowRoles の段で許可する権限(どれか1つがあれば通す)。ハンドラの中の追加の確認は
// 含まない(例: GET /api/v1/investments/:id はハンドラで ROLE_PAYMENT を必須にしている。法人・個人の PUT は
// 「ROLE_ADMIN か本人」をハンドラで見ている。法人の DELETE は、消す相手が ROLE_SUPER_ADMIN なら断る)。ROLE_ADMIN は
// 管理者なら全員が持つ。
var routeRoles = map[string][]domain.ROLE_TYPE{
	"DELETE /api/columns/:id":                                                      {domain.ROLE_OPERATOR},
	"DELETE /api/corporateusers/:id":                                               {domain.ROLE_INVESTOR},
	"DELETE /api/projects/:id":                                                     {domain.ROLE_PRODUCT},
	"DELETE /api/users/:id/consult-notes/:note_id":                                 {domain.ROLE_INVESTOR},
	"DELETE /api/v1/admin-users/:id":                                               {domain.ROLE_SUPER_ADMIN},
	"DELETE /api/v1/banners/:id":                                                   {domain.ROLE_OPERATOR},
	"DELETE /api/v1/faqs/:id":                                                      {domain.ROLE_OPERATOR},
	"DELETE /api/v1/news/:id":                                                      {domain.ROLE_OPERATOR},
	"DELETE /api/v1/policies/:id":                                                  {domain.ROLE_PRODUCT},
	"DELETE /api/v1/projects/:id/repayment-dividend-schedules/:schedule_id":        {domain.ROLE_PRODUCT},
	"GET /api/audit-response-report-output/customer-account/download":              {domain.ROLE_REPORT},
	"GET /api/audit-response-report-output/customer-account/year/:year/download":   {domain.ROLE_REPORT},
	"GET /api/audit-response-report-output/support-year":                           {domain.ROLE_REPORT},
	"GET /api/audit-response-report-output/trading-diary/download":                 {domain.ROLE_REPORT},
	"GET /api/audit-response-report-output/trading-diary/year/:year/download":      {domain.ROLE_REPORT},
	"GET /api/audit-response-report-output/transaction-record/download":            {domain.ROLE_REPORT},
	"GET /api/audit-response-report-output/transaction-record/year/:year/download": {domain.ROLE_REPORT},
	"GET /api/capital-raising-entity":                                              {domain.ROLE_PRODUCT},
	"GET /api/capital-raising-entity/:id":                                          {domain.ROLE_PRODUCT},
	"GET /api/codes":                                                               publicRoute,
	"GET /api/columns":                                                             {domain.ROLE_OPERATOR},
	"GET /api/columns/:id":                                                         {domain.ROLE_OPERATOR},
	"GET /api/contact":                                                             {domain.ROLE_INVESTOR},
	"GET /api/contact/:id":                                                         {domain.ROLE_INVESTOR},
	"GET /api/corporateusers":                                                      {domain.ROLE_INVESTOR},
	"GET /api/corporateusers/:id":                                                  {domain.ROLE_INVESTOR},
	"GET /api/corporateusers/list/anti-download":                                   {domain.ROLE_INVESTOR},
	"GET /api/corporateusers/list/download":                                        {domain.ROLE_INVESTOR},
	"GET /api/corporateusers/signup-status-count":                                  {domain.ROLE_INVESTOR},
	"GET /api/corporateusers/user-count":                                           {domain.ROLE_INVESTOR},
	"GET /api/dashboards":                                                          {domain.ROLE_ADMIN},
	"GET /api/deposit-check-data":                                                  {domain.ROLE_PAYMENT},
	"GET /api/deposit-check-data/download":                                         {domain.ROLE_PAYMENT},
	"GET /api/deposit-match-data":                                                  {domain.ROLE_PAYMENT},
	"GET /api/deposit-match-data/:id":                                              {domain.ROLE_PAYMENT},
	"GET /api/deposit-match-data/download":                                         {domain.ROLE_PAYMENT},
	"GET /api/deposit-request-lack-history/:id":                                    {domain.ROLE_PAYMENT},
	"GET /api/fundingrequesthistories":                                             {domain.ROLE_PRODUCT},
	"GET /api/fundingrequests":                                                     {domain.ROLE_PRODUCT},
	"GET /api/fundingrequests/:id":                                                 {domain.ROLE_PRODUCT},
	"GET /api/fundingrequests/exist-corp-no":                                       {domain.ROLE_PRODUCT},
	"GET /api/fundingrequests/list/download":                                       {domain.ROLE_PRODUCT},
	"GET /api/health-check":                                                        publicRoute,
	"GET /api/investor-report-output/payment-statement/list":                       {domain.ROLE_REPORT},
	"GET /api/investor-report-output/payment-statement/year/:year/download":        {domain.ROLE_REPORT},
	"GET /api/investor-report-output/transaction-balance-report/list":              {domain.ROLE_REPORT},
	"GET /api/investor-report-output/transaction-balance-report/year/:year/quarter/:quarter/download": {domain.ROLE_REPORT},
	"GET /api/japan-banks":                       {domain.ROLE_PRODUCT, domain.ROLE_INVESTOR},
	"GET /api/japan-banks/:code/branches":        {domain.ROLE_PRODUCT, domain.ROLE_INVESTOR},
	"GET /api/jobs/execute-time":                 {domain.ROLE_PAYMENT},
	"GET /api/jobs/gmo-connect-url":              {domain.ROLE_BASE},
	"GET /api/loginhistories":                    {domain.ROLE_INVESTOR},
	"GET /api/loginhistories/download":           {domain.ROLE_INVESTOR},
	"GET /api/mails/csv-form-download":           {domain.ROLE_MESSENGER},
	"GET /api/mails/receivers":                   {domain.ROLE_MESSENGER},
	"GET /api/mails/senders":                     {domain.ROLE_MESSENGER},
	"GET /api/my-number":                         {domain.ROLE_MYNUMBER},
	"GET /api/my-number/:id/check-duplicates":    {domain.ROLE_MYNUMBER},
	"GET /api/notice-msgs":                       {domain.ROLE_INVESTOR},
	"GET /api/notice-msgs/:id":                   {domain.ROLE_INVESTOR},
	"GET /api/operatorhistories":                 {domain.ROLE_SUPER_ADMIN},
	"GET /api/operatorhistories/download":        {domain.ROLE_SUPER_ADMIN},
	"GET /api/operatorhistories/log-codes":       {domain.ROLE_ADMIN},
	"GET /api/operatorhistories/menu-codes":      {domain.ROLE_ADMIN},
	"GET /api/privateusers":                      {domain.ROLE_INVESTOR},
	"GET /api/privateusers/:id":                  {domain.ROLE_INVESTOR},
	"GET /api/privateusers/:id/ekyc":             {domain.ROLE_INVESTOR},
	"GET /api/privateusers/list/anti-download":   {domain.ROLE_INVESTOR},
	"GET /api/privateusers/list/download":        {domain.ROLE_INVESTOR},
	"GET /api/privateusers/signup-status-count":  {domain.ROLE_INVESTOR},
	"GET /api/privateusers/user-count":           {domain.ROLE_INVESTOR},
	"GET /api/projectinfohistories":              {domain.ROLE_PRODUCT},
	"GET /api/projects":                          {domain.ROLE_PRODUCT},
	"GET /api/projects/:id":                      {domain.ROLE_PRODUCT},
	"GET /api/projects/:id/close-check":          {domain.ROLE_PRODUCT},
	"GET /api/projects/:id/investments":          {domain.ROLE_PRODUCT},
	"GET /api/projects/:id/investments/download": {domain.ROLE_PRODUCT},
	"GET /api/projects/check-fund-number":        {domain.ROLE_PRODUCT},
	"GET /api/projects/list/download":            {domain.ROLE_PRODUCT},
	"GET /api/qna/list":                          {domain.ROLE_INVESTOR},
	"GET /api/qna/one/:id":                       {domain.ROLE_INVESTOR},
	"GET /api/qnainfohistories":                  {domain.ROLE_INVESTOR},
	"GET /api/smsmsg":                            {domain.ROLE_MESSENGER},
	"GET /api/smsmsg/:id":                        {domain.ROLE_MESSENGER},
	"GET /api/spvinfohistories":                  {domain.ROLE_PRODUCT},
	"GET /api/spvs":                              {domain.ROLE_PRODUCT},
	"GET /api/spvs/:id":                          {domain.ROLE_PRODUCT},
	"GET /api/spvs/:id/projects":                 {domain.ROLE_PRODUCT},
	"GET /api/spvs/download":                     {domain.ROLE_PRODUCT},
	"GET /api/spvs/existscorpno":                 {domain.ROLE_PRODUCT},
	"GET /api/tax-office-report-output/payment-statement-sum/spvs/:spv_id/year/:year/download": {domain.ROLE_REPORT},
	"GET /api/tax-office-report-output/payment-statement/spvs/:spv_id/support-year":            {domain.ROLE_REPORT},
	"GET /api/tax-office-report-output/payment-statement/spvs/:spv_id/year/:year/download":     {domain.ROLE_REPORT},
	"GET /api/tax-office-report-output/spvs":                                                   {domain.ROLE_REPORT},
	"GET /api/trust-info-history":                                                              {domain.ROLE_PAYMENT},
	"GET /api/trust-infos/download":                                                            {domain.ROLE_PAYMENT},
	"GET /api/user-accounts":                                                                   {domain.ROLE_PAYMENT},
	"GET /api/user-accounts/:id":                                                               {domain.ROLE_PAYMENT},
	"GET /api/user-accounts/:id/transactions":                                                  {domain.ROLE_PAYMENT},
	"GET /api/user-accounts/download":                                                          {domain.ROLE_PAYMENT},
	"GET /api/user-transactions":                                                               {domain.ROLE_ADMIN},
	"GET /api/user-transactions/download":                                                      {domain.ROLE_ADMIN},
	"GET /api/user-virtual-accounts":                                                           {domain.ROLE_BASE},
	"GET /api/user-virtual-accounts/check-reserve":                                             {domain.ROLE_ADMIN},
	"GET /api/user-virtual-accounts/download":                                                  {domain.ROLE_BASE},
	"GET /api/userinfohistories":                                                               {domain.ROLE_INVESTOR},
	"GET /api/userinfohistories/content/:id":                                                   {domain.ROLE_INVESTOR},
	"GET /api/users":                                                                           {domain.ROLE_INVESTOR},
	"GET /api/users/:id/assets/check":                                                          {domain.ROLE_INVESTOR},
	"GET /api/users/:id/bankinfo":                                                              {domain.ROLE_INVESTOR},
	"GET /api/users/:id/consult-notes":                                                         {domain.ROLE_INVESTOR},
	"GET /api/users/:id/consult-notes/:note_id":                                                {domain.ROLE_INVESTOR},
	"GET /api/users/:id/investments":                                                           {domain.ROLE_INVESTOR},
	"GET /api/users/:id/investments/:share_id":                                                 {domain.ROLE_INVESTOR},
	"GET /api/users/:id/investments/download":                                                  {domain.ROLE_INVESTOR},
	"GET /api/users/:id/smsmsgs":                                                               {domain.ROLE_MESSENGER},
	"GET /api/v1/admin-users":                                                                  {domain.ROLE_SUPER_ADMIN},
	"GET /api/v1/admin-users/:id":                                                              {domain.ROLE_ADMIN},
	"GET /api/v1/admin-users/download":                                                         {domain.ROLE_SUPER_ADMIN},
	"GET /api/v1/banners":                                                                      {domain.ROLE_OPERATOR},
	"GET /api/v1/banners/:id":                                                                  {domain.ROLE_OPERATOR},
	"GET /api/v1/faqs":                                                                         {domain.ROLE_OPERATOR},
	"GET /api/v1/faqs/:id":                                                                     {domain.ROLE_OPERATOR},
	"GET /api/v1/investment-waiters":                                                           {domain.ROLE_PRODUCT},
	"GET /api/v1/investment-waiters/:id":                                                       {domain.ROLE_PRODUCT},
	"GET /api/v1/investment-waiters/list/download":                                             {domain.ROLE_PRODUCT},
	"GET /api/v1/investments":                                                                  {domain.ROLE_PAYMENT, domain.ROLE_INVESTOR},
	"GET /api/v1/investments/:id":                                                              {domain.ROLE_PAYMENT, domain.ROLE_INVESTOR},
	"GET /api/v1/investments/:id/transaction-histories":                                        {domain.ROLE_PAYMENT, domain.ROLE_INVESTOR},
	"GET /api/v1/investments/list/download":                                                    {domain.ROLE_PRODUCT, domain.ROLE_INVESTOR},
	"GET /api/v1/investor-report/payments":                                                     {domain.ROLE_REPORT},
	"GET /api/v1/investor-report/payments/download":                                            {domain.ROLE_REPORT},
	"GET /api/v1/investor-report/transaction-balances":                                         {domain.ROLE_REPORT},
	"GET /api/v1/investor-report/transaction-balances/download":                                {domain.ROLE_REPORT},
	"GET /api/v1/lottery/list":                                                                 {domain.ROLE_OPERATOR},
	"GET /api/v1/lottery/list/download":                                                        {domain.ROLE_OPERATOR},
	"GET /api/v1/mails":                                                                        {domain.ROLE_MESSENGER},
	"GET /api/v1/mails/:id":                                                                    {domain.ROLE_INVESTOR, domain.ROLE_MESSENGER},
	"GET /api/v1/mails/failed":                                                                 {domain.ROLE_INVESTOR, domain.ROLE_MESSENGER},
	"GET /api/v1/mails/templates/:template_type":                                               {domain.ROLE_MESSENGER},
	"GET /api/v1/news":                                                                         {domain.ROLE_OPERATOR},
	"GET /api/v1/news/:id":                                                                     {domain.ROLE_OPERATOR},
	"GET /api/v1/notice-msgs":                                                                  {domain.ROLE_INVESTOR},
	"GET /api/v1/notice-msgs/:id":                                                              {domain.ROLE_INVESTOR},
	"GET /api/v1/operatorhistories":                                                            {domain.ROLE_SUPER_ADMIN},
	"GET /api/v1/operatorhistories/download":                                                   {domain.ROLE_SUPER_ADMIN},
	"GET /api/v1/operatorhistories/log-codes":                                                  {domain.ROLE_ADMIN},
	"GET /api/v1/operatorhistories/menu-codes":                                                 {domain.ROLE_ADMIN},
	"GET /api/v1/policies":                                                                     {domain.ROLE_PRODUCT},
	"GET /api/v1/policies/:id":                                                                 {domain.ROLE_PRODUCT},
	"GET /api/v1/policies/:id/check-label":                                                     {domain.ROLE_PRODUCT},
	"GET /api/v1/policies/:id/histories":                                                       {domain.ROLE_PRODUCT},
	"GET /api/v1/projects/:id/lottery-applications/bulk-lottery/download":                      {domain.ROLE_PRODUCT},
	"GET /api/v1/projects/:id/repayment-dividend-schedules/info":                               {domain.ROLE_PRODUCT},
	"GET /api/v1/projects/:id/repayment-dividend-schedules/table":                              {domain.ROLE_PRODUCT},
	"GET /api/v1/projects/:id/repayment-dividend-schedules/table-csv":                          {domain.ROLE_PRODUCT},
	"GET /api/v1/projects/:id/statistics":                                                      {domain.ROLE_PRODUCT},
	"GET /api/v1/qna/changelogs":                                                               {domain.ROLE_INVESTOR},
	"GET /api/v1/users/:id/mails":                                                              {domain.ROLE_INVESTOR},
	"GET /api/v1/withdrawals":                                                                  {domain.ROLE_PAYMENT},
	"GET /api/v1/withdrawals/:id":                                                              {domain.ROLE_PAYMENT},
	"GET /api/v1/withdrawals/closed":                                                           {domain.ROLE_PAYMENT},
	"GET /api/v1/withdrawals/closed/:id":                                                       {domain.ROLE_PAYMENT},
	"GET /api/v1/withdrawals/closed/:id/download":                                              {domain.ROLE_PAYMENT},
	"GET /api/withdrawal":                                                                      {domain.ROLE_PAYMENT},
	"GET /api/withdrawal/:id":                                                                  {domain.ROLE_PAYMENT},
	"GET /api/withdrawal/closed":                                                               {domain.ROLE_PAYMENT},
	"GET /api/withdrawal/closed/:id":                                                           {domain.ROLE_PAYMENT},
	"GET /api/withdrawal/closed/:id/download":                                                  {domain.ROLE_PAYMENT},
	"GET /doc":                                                          publicRoute,
	"GET /doc/swagger.json":                                             publicRoute,
	"PATCH /api/v1/withdrawals/:id":                                     {domain.ROLE_PAYMENT},
	"POST /api/auth/update-token":                                       publicRoute,
	"POST /api/capital-raising-entity":                                  {domain.ROLE_PRODUCT},
	"POST /api/columns":                                                 {domain.ROLE_OPERATOR},
	"POST /api/jobs/execute-deposit-check-api":                          {domain.ROLE_PAYMENT},
	"POST /api/jobs/execute-deposit-process-api":                        {domain.ROLE_PAYMENT},
	"POST /api/mails/receivers/count":                                   {domain.ROLE_MESSENGER},
	"POST /api/mails/receivers/count-csv":                               {domain.ROLE_MESSENGER},
	"POST /api/mails/send":                                              {domain.ROLE_MESSENGER},
	"POST /api/mails/send/csv":                                          {domain.ROLE_MESSENGER},
	"POST /api/notice-msgs/send":                                        {domain.ROLE_INVESTOR},
	"POST /api/notice-msgs/send/newsletter":                             {domain.ROLE_INVESTOR},
	"POST /api/notice-msgs/send/projects":                               {domain.ROLE_INVESTOR},
	"POST /api/notice-msgs/send/user-type":                              {domain.ROLE_INVESTOR},
	"POST /api/projects":                                                {domain.ROLE_PRODUCT},
	"POST /api/qna/:id/answer":                                          {domain.ROLE_INVESTOR},
	"POST /api/referral-codes/issue":                                    {domain.ROLE_OPERATOR},
	"POST /api/smsmsg":                                                  {domain.ROLE_MESSENGER},
	"POST /api/smsmsg/member":                                           {domain.ROLE_MESSENGER},
	"POST /api/spvs":                                                    {domain.ROLE_PRODUCT},
	"POST /api/storage/upload":                                          {domain.ROLE_ADMIN},
	"POST /api/user-virtual-accounts/upload-equity-account-csv":         {domain.ROLE_REPORT},
	"POST /api/user-virtual-accounts/upload-loan-account-csv":           {domain.ROLE_REPORT},
	"POST /api/users/:id/antisocial-check-update":                       {domain.ROLE_INVESTOR},
	"POST /api/users/:id/consult-notes":                                 {domain.ROLE_INVESTOR},
	"POST /api/users/:id/link-virtual-account":                          {domain.ROLE_INVESTOR},
	"POST /api/users/passwordresetrequest":                              {domain.ROLE_ADMIN},
	"POST /api/v1/admin-users":                                          {domain.ROLE_SUPER_ADMIN},
	"POST /api/v1/admin-users/login":                                    publicRoute,
	"POST /api/v1/admin-users/logout":                                   {domain.ROLE_ADMIN},
	"POST /api/v1/banners":                                              {domain.ROLE_OPERATOR},
	"POST /api/v1/faqs":                                                 {domain.ROLE_OPERATOR},
	"POST /api/v1/files/pre-signed-urls":                                {domain.ROLE_OPERATOR},
	"POST /api/v1/investment-waiters/:id/cancel":                        {domain.ROLE_PRODUCT},
	"POST /api/v1/investment-waiters/:id/investment-propose":            {domain.ROLE_PRODUCT},
	"POST /api/v1/investment-waiters/:id/investment-request":            {domain.ROLE_PRODUCT},
	"POST /api/v1/investment-waiters/:id/propose":                       {domain.ROLE_PRODUCT},
	"POST /api/v1/lottery/applications/:lottery_application_id/cancel":  {domain.ROLE_OPERATOR},
	"POST /api/v1/lottery/applications/:lottery_application_id/lose":    {domain.ROLE_OPERATOR},
	"POST /api/v1/lottery/applications/:lottery_application_id/win":     {domain.ROLE_OPERATOR},
	"POST /api/v1/lottery/applications/:project_id/csv-upload":          {domain.ROLE_OPERATOR},
	"POST /api/v1/news":                                                 {domain.ROLE_OPERATOR},
	"POST /api/v1/notice-msgs/send":                                     {domain.ROLE_INVESTOR},
	"POST /api/v1/notice-msgs/send/newsletter":                          {domain.ROLE_INVESTOR},
	"POST /api/v1/notice-msgs/send/projects":                            {domain.ROLE_INVESTOR},
	"POST /api/v1/notice-msgs/send/user-type":                           {domain.ROLE_INVESTOR},
	"POST /api/v1/projects/:id/execute-investment":                      {domain.ROLE_PRODUCT},
	"POST /api/v1/projects/:id/repayment-dividend-schedules":            {domain.ROLE_PRODUCT},
	"POST /api/v1/users/filter/csv":                                     {domain.ROLE_INVESTOR},
	"POST /api/v1/withdrawals":                                          {domain.ROLE_PAYMENT},
	"POST /api/v1/withdrawals/:id/transfer":                             {domain.ROLE_PAYMENT},
	"POST /api/v1/withdrawals/closed/:id/transfer":                      {domain.ROLE_PAYMENT},
	"POST /api/withdrawal/:id/transfer":                                 {domain.ROLE_PAYMENT},
	"POST /api/withdrawal/closed/:id/transfer":                          {domain.ROLE_PAYMENT},
	"PUT /api/capital-raising-entity/:id":                               {domain.ROLE_PRODUCT},
	"PUT /api/columns/:id":                                              {domain.ROLE_OPERATOR},
	"PUT /api/corporateusers/:id":                                       {domain.ROLE_OPERATOR, domain.ROLE_INVESTOR},
	"PUT /api/corporateusers/:id/auditstatus":                           {domain.ROLE_INVESTOR},
	"PUT /api/corporateusers/:id/blockinvest/:block_invest":             {domain.ROLE_INVESTOR},
	"PUT /api/corporateusers/:id/investor-type":                         {domain.ROLE_INVESTOR},
	"PUT /api/fundingrequests/:id":                                      {domain.ROLE_PRODUCT},
	"PUT /api/investor-report-output/payment-statement/update":          {domain.ROLE_REPORT},
	"PUT /api/investor-report-output/transaction-balance-report/update": {domain.ROLE_REPORT},
	"PUT /api/my-number/:id":                                            {domain.ROLE_MYNUMBER},
	"PUT /api/privateusers/:id":                                         {domain.ROLE_OPERATOR, domain.ROLE_INVESTOR},
	"PUT /api/privateusers/:id/auditstatus":                             {domain.ROLE_INVESTOR},
	"PUT /api/privateusers/:id/blockinvest/:block_invest":               {domain.ROLE_INVESTOR},
	"PUT /api/privateusers/:id/investor-type":                           {domain.ROLE_INVESTOR},
	"PUT /api/projects/:id":                                             {domain.ROLE_PRODUCT},
	"PUT /api/projects/:id/project-status/:project_status":              {domain.ROLE_PRODUCT},
	"PUT /api/qna/:id/answer/:answerId":                                 {domain.ROLE_INVESTOR},
	"PUT /api/spvs/:id":                                                 {domain.ROLE_PRODUCT},
	"PUT /api/users/:id/consult-notes/:note_id":                         {domain.ROLE_INVESTOR},
	"PUT /api/v1/admin-users/:id":                                       {domain.ROLE_SUPER_ADMIN},
	"PUT /api/v1/banners/:id":                                           {domain.ROLE_OPERATOR},
	"PUT /api/v1/banners/rolling-time":                                  {domain.ROLE_OPERATOR},
	"PUT /api/v1/faqs/:id":                                              {domain.ROLE_OPERATOR},
	"PUT /api/v1/investment-waiters/:id/status-update":                  {domain.ROLE_PRODUCT},
	"PUT /api/v1/investments/:id/invest-status":                         {domain.ROLE_PAYMENT, domain.ROLE_INVESTOR},
	"PUT /api/v1/investor-report/payments":                              {domain.ROLE_REPORT},
	"PUT /api/v1/investor-report/transaction-balances":                  {domain.ROLE_REPORT},
	"PUT /api/v1/news/:id":                                              {domain.ROLE_OPERATOR},
	"PUT /api/v1/policies/:id":                                          {domain.ROLE_PRODUCT},
	"PUT /api/v1/policies/:id/histories/:history_id":                    {domain.ROLE_PRODUCT},
	"PUT /api/v1/withdrawals/closed":                                    {domain.ROLE_PAYMENT},
	"PUT /api/withdrawal/:id":                                           {domain.ROLE_PAYMENT},
	"PUT /api/withdrawal/closed":                                        {domain.ROLE_PAYMENT},
}

// stagingOnlyRoutes は、開発環境・staging(cfg.IsDevelopment())のときだけ登録するルート。
var stagingOnlyRoutes = map[string][]domain.ROLE_TYPE{
	"PATCH /api/v1/maintenance/deposit-requests":             {domain.ROLE_SUPER_ADMIN},
	"PATCH /api/v1/maintenance/projects":                     {domain.ROLE_SUPER_ADMIN},
	"PATCH /api/v1/maintenance/repayment-dividend-schedules": {domain.ROLE_SUPER_ADMIN},
}

// adminRoles は、権限ごとの試験で1つずつ ROLE_ADMIN に足す権限。
var adminRoles = []domain.ROLE_TYPE{
	domain.ROLE_ADMIN,
	domain.ROLE_OPERATOR,
	domain.ROLE_PRODUCT,
	domain.ROLE_PAYMENT,
	domain.ROLE_INVESTOR,
	domain.ROLE_MESSENGER,
	domain.ROLE_BASE,
	domain.ROLE_REPORT,
	domain.ROLE_MYNUMBER,
	domain.ROLE_SUPER_ADMIN,
}

var routeRolesEnvs = []string{"prod", "stg"}

func expectedRouteRoles(env string) map[string][]domain.ROLE_TYPE {
	expected := make(map[string][]domain.ROLE_TYPE, len(routeRoles)+len(stagingOnlyRoutes))
	for k, v := range routeRoles {
		expected[k] = v
	}
	cfg := &config.Config{}
	cfg.Env = env
	if cfg.IsDevelopment() {
		for k, v := range stagingOnlyRoutes {
			expected[k] = v
		}
	}
	return expected
}

func isPublic(roles []domain.ROLE_TYPE) bool {
	return len(roles) == 1 && roles[0] == publicRoute[0]
}

func routeKey(r recordedRoute) string {
	return r.method + " " + r.path
}

func TestRouteRoles_登録されたルートと表が一致する(t *testing.T) {
	for _, env := range routeRolesEnvs {
		t.Run(env, func(t *testing.T) {
			routes, _ := recordRoutes(t, env)
			expected := expectedRouteRoles(env)

			seen := map[string]int{}
			for _, r := range routes {
				key := routeKey(r)
				seen[key]++
				assert.Equalf(t, 1, seen[key], "%s が二重に登録されている(echo は後の登録で黙って上書きする)", key)
				_, ok := expected[key]
				assert.Truef(t, ok, "%s が表に無い。routeRoles(開発環境・staging だけのルートは stagingOnlyRoutes)に許可する権限を書く", key)
			}
			for key := range expected {
				assert.Containsf(t, seen, key, "%s は表にあるが登録されていない", key)
			}
		})
	}
}

func TestRouteRoles_会員のトークンでは公開以外のルートを呼べない(t *testing.T) {
	// MIR-1407 のあと、会員のトークンには ROLE_USER しか入らない前提を試す。
	for _, env := range routeRolesEnvs {
		t.Run(env, func(t *testing.T) {
			routes, cfg := recordRoutes(t, env)
			e := replayEcho(cfg, routes)
			expected := expectedRouteRoles(env)
			userToken := tokenWithRoles(t, domain.ROLE_USER)

			for _, r := range routes {
				if isPublic(expected[routeKey(r)]) {
					continue
				}
				code, _ := call(e, r.method, r.path, userToken)
				assert.Equalf(t, http.StatusForbidden, code, "%s", routeKey(r))
			}
		})
	}
}

func TestRouteRoles_権限ごとに表のとおり通す(t *testing.T) {
	for _, env := range routeRolesEnvs {
		t.Run(env, func(t *testing.T) {
			routes, cfg := recordRoutes(t, env)
			e := replayEcho(cfg, routes)
			expected := expectedRouteRoles(env)

			tokens := make(map[domain.ROLE_TYPE]string, len(adminRoles))
			for _, role := range adminRoles {
				// 本番の管理者は全員 ROLE_ADMIN を持つので、常に入れる。
				tokens[role] = tokenWithRoles(t, domain.ROLE_ADMIN, role)
			}

			for _, r := range routes {
				key := routeKey(r)
				allowed, ok := expected[key]
				if !ok || isPublic(allowed) {
					continue
				}
				for _, role := range adminRoles {
					want := containsRole(allowed, domain.ROLE_ADMIN) || containsRole(allowed, role)
					code, body := call(e, r.method, r.path, tokens[role])
					if want {
						// 本文は仮のハンドラが返す登録したパス。別のルートに当たっていないことも確かめる。
						assert.Equalf(t, http.StatusOK, code, "%s を %s で呼べるはず", key, role)
						assert.Equalf(t, r.path, body, "%s の URL が別のルートに当たっている", key)
					} else {
						assert.Equalf(t, http.StatusForbidden, code, "%s を %s で呼べてはいけない", key, role)
					}
				}
			}
		})
	}
}

func TestRouteRoles_トークン無しでは公開のルートしか呼べない(t *testing.T) {
	for _, env := range routeRolesEnvs {
		t.Run(env, func(t *testing.T) {
			routes, cfg := recordRoutes(t, env)
			e := replayEcho(cfg, routes)
			expected := expectedRouteRoles(env)

			for _, r := range routes {
				key := routeKey(r)
				code, body := call(e, r.method, r.path, "")
				if isPublic(expected[key]) {
					assert.Equalf(t, http.StatusOK, code, "%s は公開のはず", key)
					assert.Equalf(t, r.path, body, "%s の URL が別のルートに当たっている", key)
				} else {
					assert.Equalf(t, http.StatusUnauthorized, code, "%s はトークン無しで 401 のはず", key)
				}
			}
		})
	}
}

func TestJwtSkipYml_公開のルートだけが載っている(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(pathutils.GetPathOfAssets(), "tabledata/admin-api/jwt_skip.yml"))
	require.NoError(t, err)
	var parsed struct {
		JwtSkipData []struct {
			Path   string `yaml:"path"`
			Method string `yaml:"method"`
		} `yaml:"jwt_skip_data"`
	}
	require.NoError(t, yaml.Unmarshal(data, &parsed))
	require.NotEmpty(t, parsed.JwtSkipData)

	// /api/swagger* は開発環境・staging のサーバーの起動時にだけ登録され、組み立てには出てこない。
	const swaggerUI = "/api/swagger*"
	knownMethods := map[string]bool{
		http.MethodGet: true, http.MethodPost: true, http.MethodPut: true,
		http.MethodPatch: true, http.MethodDelete: true,
	}

	for _, env := range routeRolesEnvs {
		t.Run(env, func(t *testing.T) {
			routes, _ := recordRoutes(t, env)
			expected := expectedRouteRoles(env)
			methodsByPath := map[string][]string{}
			for _, r := range routes {
				methodsByPath[r.path] = append(methodsByPath[r.path], r.method)
			}

			seenPath := map[string]bool{}
			for _, skip := range parsed.JwtSkipData {
				assert.Falsef(t, seenPath[skip.Path], "%s が2行ある(後ろの行が前の行を上書きする)", skip.Path)
				seenPath[skip.Path] = true
				if skip.Path == swaggerUI {
					continue
				}
				registered, ok := methodsByPath[skip.Path]
				if !assert.Truef(t, ok, "%s は登録されていない(存在しないルートを載せない)", skip.Path) {
					continue
				}
				// method が空なら、そのパスの全メソッドを飛ばす。
				methods := registered
				if skip.Method != "" {
					methods = strings.Split(skip.Method, ",")
				}
				for _, m := range methods {
					assert.Truef(t, knownMethods[m], "%s の method %q が不正(空白を入れない・既知のメソッドだけ)", skip.Path, m)
					key := m + " " + skip.Path
					assert.Truef(t, isPublic(expected[key]), "%s は表で公開ではないのに、トークンの検査を飛ばしている", key)
				}
			}
		})
	}
}

func TestAdminRoles_domainの権限をすべて含む(t *testing.T) {
	// adminRoles に無い権限で守られたルートは、権限ごとの試験で「許可された権限で呼べるか」を確かめられない。
	// domain の全ファイルから ROLE_TYPE の定数(型を書いたものと ROLE_TYPE("...") の形)の値を集め、会員の ROLE_USER
	// を除いて adminRoles と一致することを確かめる。
	files, err := filepath.Glob(filepath.Join(filepath.Dir(pathutils.GetPathOfAssets()), "domain", "*.go"))
	require.NoError(t, err)

	var declared []string
	fset := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		require.NoError(t, err)
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				typed := isRoleTypeIdent(vs.Type)
				for _, value := range vs.Values {
					lit := value
					if call, ok := value.(*ast.CallExpr); ok && isRoleTypeIdent(call.Fun) && len(call.Args) == 1 {
						lit = call.Args[0]
					} else if !typed {
						continue
					}
					basic, ok := lit.(*ast.BasicLit)
					require.Truef(t, ok && basic.Kind == token.STRING, "%s: ROLE_TYPE の定数は文字列で書く", fset.Position(value.Pos()))
					role, err := strconv.Unquote(basic.Value)
					require.NoError(t, err)
					if role != string(domain.ROLE_USER) {
						declared = append(declared, role)
					}
				}
			}
		}
	}
	require.NotEmpty(t, declared)

	listed := make([]string, 0, len(adminRoles))
	for _, r := range adminRoles {
		listed = append(listed, string(r))
	}
	assert.ElementsMatch(t, declared, listed)
}

func isRoleTypeIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "ROLE_TYPE"
}

func TestConfigureEcho_infraはトークンの検査より前に動く(t *testing.T) {
	// 本番では infra に Kafka へのアクセスログと DB が入る。トークンの検査より後ろへ動かすと、401 の要求がアクセス
	// ログに残らなくなるので、順番をここで固定する。
	cfg := &config.Config{}
	cfg.Env = "prod"
	cfg.Jwt.Secret = routeRolesTestSecret

	ran := false
	marker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ran = true
			return next(c)
		}
	}
	e := echo.New()
	configureEcho(e, cfg, marker)
	e.GET("/api/needs-token", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	code, _ := call(e, http.MethodGet, "/api/needs-token", "")
	assert.Equal(t, http.StatusUnauthorized, code)
	assert.True(t, ran, "トークン無しで 401 になる要求でも、infra のミドルウェアが動くこと")
}

func containsRole(roles []domain.ROLE_TYPE, role domain.ROLE_TYPE) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

const routeRolesTestSecret = "route-roles-test-secret"

// recordedRoute は、本番と同じ組み立て(appOptions)で登録されたルートと、そのルートに付いたミドルウェア(グループと
// ルートの両方。admin-api では AllowRoles だけ)。
type recordedRoute struct {
	method      string
	path        string
	middlewares []echo.MiddlewareFunc
}

// recordRoutes は appOptions で echo を組み立て、登録されたルートを OnAddRouteHandler で全部記録する。DB や外部の
// 接続は持たない。同じメソッド・パスの二重登録は、echo が黙って上書きするので、ここで数える。
func recordRoutes(t *testing.T, env string) ([]recordedRoute, *config.Config) {
	t.Helper()
	cfg := &config.Config{}
	cfg.Env = env
	cfg.Jwt.Secret = routeRolesTestSecret

	e := echo.New()
	var routes []recordedRoute
	e.OnAddRouteHandler = func(_ string, route echo.Route, _ echo.HandlerFunc, middlewares []echo.MiddlewareFunc) {
		if route.Method == echo.RouteNotFound {
			return
		}
		routes = append(routes, recordedRoute{method: route.Method, path: route.Path, middlewares: middlewares})
	}
	configureEcho(e, cfg)

	app := fx.New(appOptions([]interface{}{cfg, e}), fx.NopLogger)
	require.NoError(t, app.Err())
	return routes, cfg
}

// replayEcho は、記録したルートを、同じグローバルなミドルウェア(configureEcho)と同じルートのミドルウェアで、
// 決まった応答(200 と登録したパス)を返す仮のハンドラに登録し直す。ハンドラを動かさないので、ハンドラの中の
// 追加の確認や DB・外部の通信に左右されない。
func replayEcho(cfg *config.Config, routes []recordedRoute) *echo.Echo {
	e := echo.New()
	configureEcho(e, cfg)
	for _, r := range routes {
		e.Add(r.method, r.path, func(c echo.Context) error {
			return c.String(http.StatusOK, c.Path())
		}, r.middlewares...)
	}
	return e
}

var pathParam = regexp.MustCompile(`:[^/]+`)

// concretePath は、ルートの雛形のパラメータを具体的な値にした URL。
func concretePath(path string) string {
	return pathParam.ReplaceAllString(path, "1")
}

func tokenWithRoles(t *testing.T, roles ...domain.ROLE_TYPE) string {
	t.Helper()
	roleClaims := make([]string, 0, len(roles))
	for _, r := range roles {
		roleClaims = append(roleClaims, string(r))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   1,
		"role": roleClaims,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(routeRolesTestSecret))
	require.NoError(t, err)
	return signed
}

// call は、トークン(空なら付けない)を付けてルートを呼び、状態コードと本文を返す。
func call(e *echo.Echo, method, path, token string) (int, string) {
	req := httptest.NewRequest(method, concretePath(path), nil)
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
