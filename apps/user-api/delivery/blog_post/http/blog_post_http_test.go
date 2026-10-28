package http

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"

	pkgtest "github.com/illuminarean/energy-share-core-api/delivery/testing"
	cdomain "github.com/illuminarean/energy-shares-backend-repo/clean-arch-domain"
	"github.com/illuminarean/energy-shares-backend-repo/pkg/validator"
	udomain "github.com/illuminarean/energyshares-user-api/domain"
	"github.com/illuminarean/energyshares-user-api/domain/mocks"
)

func Test_blogPostHttpDelivery_GetAll(t *testing.T) {
	tests := []struct {
		name                string
		echoCtx             func(e *echo.Echo) echo.Context
		blogpostUserUsecase func() *mocks.BlogPostUsecase
		wantErr             bool
	}{
		{
			name: "GET /v1/blog-posts",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/v1/blog-posts",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"limit":  "0",
					"offset": "0",
				},
			},
			),
			blogpostUserUsecase: func() *mocks.BlogPostUsecase {
				m := &mocks.BlogPostUsecase{}
				m.On("GetAll", mock.Anything, udomain.BlogPostGetAllRequest{
					Offset: 0,
					Limit:  0,
				}).Return(int64(1), udomain.BlogPostGetAllResponse{
					TotalCount: 1,
					Items: []udomain.BlogPostResponse{
						{
							Id:       "018fdc2b-2199-7217-9e58-c87b53ab4b0b",
							Title:    "COP28で新展開！ 脱炭素化と再生可能エネルギーへの投資",
							Link:     "https://stg.energyshares.jp/study/cop28-decarbonation/",
							Data:     "\n<!--kg-card-begin: html-->\n\n<p>2023年11月30日から約2週間にわたって、約200か国が参加した、国連の気候変動対策会議「COP28」が開催されました。再生可能エネルギーの目標も議論され、脱炭素化への動きはますます加速していくと思います。この記事では、COP28や世界的な脱炭素化、再生可能エネルギーの動向について、ご紹介していきます！</p>\n\n\n\n\n\n\n\n<p class=\"has-text-align-left\">COP（コップ）は、締約国会議（Conference of the Parties）の略で、条約を結んだ国々による会議のこと。1992年に国連で採択された気候変動枠組条約を結んだ国や地域が参加する、COP（国連気候変動枠組条約締約国会議）は28回目になり、COP28がドバイで開催。</p>\n\n\n\n<p></p>\n\n\n\n<p><strong>温室効果ガスの濃度を安定させることや、地球温暖化がもたらす様々な悪影響を防ぐことが目標</strong>の会議です。その目標を実現するために、国際的なルールづくりが行われます。</p>\n\n\n\n<p></p>\n\n\n\n<div class=\"wp-block-vk-blocks-heading\" id=\"midashi2\"><div class=\"vk_heading vk_heading-style-default\"><h2 class=\"vk_heading_title vk_heading_title-style-default\"><span id=\"i\"><span>脱炭素化の国際的な動き</span></span></h2></div></div>\n\n\n\n<p>COP21で採択された「パリ協定」では、下記のような世界共通の長期目標が掲げられました。</p>\n\n\n\n<p></p>\n\n\n\n<p><strong>「<span data-color=\"#fffd6b\" style=\"background: linear-gradient(transparent 60%,rgba(255, 253, 107, 0.7) 0);\" class=\"vk_highlighter\">世界の平均気温上昇を産業革命以前に比べて2度よりも十分低く保ち、1.5度に抑える努力を追求する</span>」</strong></p>\n\n\n\n<p></p>\n\n\n\n<p>これらは、「<strong><span data-color=\"#fffd6b\" style=\"background: linear-gradient(transparent 60%,rgba(255, 253, 107, 0.7) 0);\" class=\"vk_highlighter\">2度目標</span></strong>」「<span data-color=\"#fffd6b\" style=\"background: linear-gradient(transparent 60%,rgba(255, 253, 107, 0.7) 0);\" class=\"vk_highlighter\"><strong>1.5度目標</strong></span>」と呼ばれています。パリ協定を受けて、先進国を中心にカーボンニュートラル（温室効果ガスの排出量と吸収量の差し引きをゼロにすること）の実現を表明しています。</p>\n\n\n\n<p>COP28では、「グローバル・ストックテイク（GST）」の実施が注目されました。GSTは、パリ協定に基づく気候変動対策の目標をどのくらい達成したかを5年ごとに評価する仕組みです。</p>\n\n\n\n<p>目標に向けて、行動と支援が必要であることが強調されました。<strong>2030年までに再生可能エネルギーの発電量を世界全体で3倍にする</strong>などが掲げられ、脱炭素化への動きは世界中で加速していくと思います。</p>\n\n\n\n<p><span data-fontsize=\"15px\" style=\"font-size: 15px;\" class=\"vk_inline-font-size\"><span style=\"text-decoration: underline;\">参考：<a href=\"https://www.enecho.meti.go.jp/about/special/johoteikyo/cop28_01.html?ref=stg.energyshares.jp\">経済産業省 資源エネルギー庁</a></span></span></p>\n\n\n\n<p></p>\n\n\n\n<p>2024年の気候変動対策会議「COP29」は、11月11日〜22日に旧ソビエトのアゼルバイジャンで開かれることになりました。次回の会議に向けて、世界の取り組みに注目したいです。</p>\n\n\n\n<p><span style=\"text-decoration: underline;\"><span data-fontsize=\"15px\" style=\"font-size: 15px;\" class=\"vk_inline-font-size\">参考：<a href=\"https://www3.nhk.or.jp/news/html/20231212/k10014286031000.html?ref=stg.energyshares.jp\">NHKニュース</a></span></span></p>\n\n\n\n<p></p>\n\n\n\n<figure class=\"wp-block-image size-full\"><img loading=\"lazy\" width=\"940\" height=\"529\" src=\"https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093733/image-3.png\" alt class=\"wp-image-1334\" srcset=\"https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093733/image-3.png 940w, https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093733/image-3-300x169.png 300w, https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093733/image-3-768x432.png 768w\" sizes=\"(max-width: 940px) 100vw, 940px\"></figure>\n\n\n\n<p></p>\n\n\n\n<h2 id=\"midashi3\"><span id=\"i-2\">再生可能エネルギーへの投資</span></h2>\n\n\n\n<p>これまでの化石エネルギー（石油・石炭など）から脱却し、再生可能エネルギー（太陽光・風力・水力・地熱など）に対するニーズは高まっています。</p>\n\n\n\n<p>環境にやさしく、枯渇する心配がないため、新しいエネルギー源として注目されており、投資としても大きなチャンスがあります。</p>\n\n\n\n<p>アメリカでは再生可能エネルギーへの投資が増えており、再生可能エネルギーに関する投資には追い風が吹いているでしょう。</p>\n\n\n\n<p><span data-fontsize=\"15px\" style=\"font-size: 15px;\" class=\"vk_inline-font-size\"><span style=\"text-decoration: underline;\">参考：<a href=\"https://xtech.nikkei.com/atcl/nxt/column/18/02443/090400016/?ref=stg.energyshares.jp\">日経XTECH</a></span></span></p>\n\n\n\n<p></p>\n\n\n\n<h2 id=\"midashi4\"><span id=\"i-3\">投資の利点とリスク</span></h2>\n\n\n\n<p>再生可能エネルギーの需要は今後も伸びる可能性が高く、投資が社会貢献に繋がっていくことが考えられます。環境問題に対する取り組みをサポートでき、投資によって社会をより良くしていくことができます。</p>\n\n\n\n<p><strong>リスクとしては、災害や天候が収益に影響すること</strong>です。再生可能エネルギーの中でもシェアを大きく占めている太陽光発電は、悪天候が続くと収益が落ちてしまいます。</p>\n\n\n\n<p>国の政策にも左右される可能性があるため、投資をする際は継続的に事業を行う意思がある事業者かを見極める必要があります。</p>\n\n\n\n<p></p>\n\n\n\n<figure class=\"wp-block-image size-full\"><img loading=\"lazy\" width=\"940\" height=\"627\" src=\"https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093854/image-4.png\" alt class=\"wp-image-1335\" srcset=\"https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093854/image-4.png 940w, https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093854/image-4-300x200.png 300w, https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16093854/image-4-768x512.png 768w\" sizes=\"(max-width: 940px) 100vw, 940px\"></figure>\n\n\n\n<p></p>\n\n\n\n<h2 id=\"midashi5\"><span id=\"i-4\">持続可能な未来への投資の重要性</span></h2>\n\n\n\n<p>再生可能エネルギーは、持続可能な未来を実現するために成長性や将来性が見込まれている分野です。気候変動対策だけでなく、戦争などによる世界情勢の変化から、自国でのエネルギー供給に関しても重点が置かれるようになりました。</p>\n\n\n\n<p><strong>化石燃料の依存を減らし、再生可能エネルギーを国内で発電するインフラ整備が不可欠になっています。</strong>再生可能エネルギーへの投資は、これからの社会づくりに重要な役割を果たすと考えられるでしょう。</p>\n\n\n\n<p></p>\n\n\n\n<h2 id=\"midashi6\"><span id=\"i-5\">まとめ</span></h2>\n\n\n\n<p>COP28は、温室効果ガスの濃度を安定させることや、地球温暖化がもたらす様々な悪影響を防ぐことが目標の世界会議です。気候変動対策の目標をどのくらい達成したかを評価する仕組みが、今回実施されました。</p>\n\n\n\n<p><strong><span data-color=\"#fffd6b\" style=\"background: linear-gradient(transparent 60%,rgba(255, 253, 107, 0.7) 0);\" class=\"vk_highlighter\">2030年までに再生可能エネルギーの発電量を世界全体で3倍にする</span></strong>などが掲げられ、脱炭素化への動きは世界中で加速していくと思います。</p>\n\n\n\n<p>再生可能エネルギーの需要は今後も伸びる可能性が高く、投資が社会貢献に繋がっていくことが考えられます。</p>\n\n\n\n<p></p>\n\n\n\n<p>これからの社会づくりに重要な役割を果たす、再生可能エネルギーの投資。</p>\n\n\n\n<p>環境問題への取り組みをサポートできる投資で、より良い社会をつくっていきましょう！</p>\n\n\n\n<p></p>\n\n\n\n<figure class=\"wp-block-image size-full\"><img loading=\"lazy\" width=\"940\" height=\"529\" src=\"https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16094008/image-5.png\" alt class=\"wp-image-1336\" srcset=\"https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16094008/image-5.png 940w, https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16094008/image-5-300x169.png 300w, https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16094008/image-5-768x432.png 768w\" sizes=\"(max-width: 940px) 100vw, 940px\"></figure>\n\n\n\n<p></p>\n\n\n\n<p><strong>▼再生可能エネルギーのESG投資プラットフォーム「エナジーシェアーズ」</strong><br><a href=\"https://energyshares.jp/?ref=stg.energyshares.jp\">https://energyshares.jp/</a></p>\n\n<!--kg-card-end: html-->\n",
							Category: "",
							Img:      "https://prod-wp-cdn.energyshares.jp/wp-content/uploads/2024/01/16094903/image-6.png",
							Date:     "2024-01-01",
						},
					},
				}, nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "GET /v1/blog-posts",
			echoCtx: pkgtest.RequestContext(pkgtest.Request{
				Path:   "/v1/blog-posts",
				Method: http.MethodGet,
				Header: pkgtest.Header{
					Headers:   echo.HeaderContentType,
					MIMETypes: echo.MIMEApplicationJSON,
				},
				Params: map[string]string{
					"limit":  "0",
					"offset": "0",
				},
			},
			),
			blogpostUserUsecase: func() *mocks.BlogPostUsecase {
				m := &mocks.BlogPostUsecase{}
				m.On("GetAll", mock.Anything, udomain.BlogPostGetAllRequest{
					Offset: 0,
					Limit:  0,
				}).Return(int64(0), udomain.BlogPostGetAllResponse{
					TotalCount: 0,
					Items:      []udomain.BlogPostResponse{},
				}, cdomain.ErrItemNotFound)
				return m
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := blogPostHttpDelivery{
				blogPostUsecase: tt.blogpostUserUsecase(),
			}
			e := echo.New()
			e.Validator = validator.NewCustomValidator()
			if err := delivery.GetAll(tt.echoCtx(e)); (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
