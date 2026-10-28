# Energy Shares JP User APIサーバー

## Getting Started && Config
 ```
  go run main.go run --env local
  YAML FILE : config/config.local.yml 
  
  // 開発環境
  go run main.go run --env dev
  YAML FILE : config/config.dev.yml 
  
  // STG環境
  go run main.go run --env stg-qa-phase2
  YAML FILE : config/config.stg.yml 
  
   // PROD環境
  go run main.go run --env prod
  YAML FILE : config/config.prod.yml 
```

## Test
```
$ go test -v ./...
// または 
$ make test
```

## ポート修正
```
// config/config.*.ymlファイルでhttpPortを修正
userConfig:
  httpPort: 8081
  service: energy-shares-dev-user-api
```

## URL ##
Visit http://127.0.0.1:8080/

API Docs http://127.0.0.1:8080/doc


## GMO-Aozora銀行テスト用アカウント
- ID: 7260008487
- Password cq52s6k6
- BranchCode: 101
- ESKKAccount: 0001465
