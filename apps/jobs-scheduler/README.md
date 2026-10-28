# energyshares-job-scheduler

## go mod Private Repository

go mod は、バージョンを指定しない場合、デフォルトでリリース済みのバージョンをダウンロードします。  
この挙動を無視して、自分が作業中のブランチのライブラリをダウンロードしたい場合は、以下のコマンドを実行すると go.mod ファイルに
希望するバージョンのライブラリが登録・ダウンロードされます。
```
GOSUMDB=off go get -v github.com/illuminarean/energy-share-core-api@[ダウンロードするブランチのバージョン]
```

## EnergyShares Batch List

* バッチ関数を追加または修正した際は、以下のドキュメントを更新してください  
  https://docs.google.com/spreadsheets/d/1eKQbZG7L8J42TT318_Ie8EUfVTQgAmBU4NffhwOHdfg/edit#gid=0

# He

## Cron

- https://pkg.go.dev/github.com/robfig/cron/v3

### CRON Expression Format ¶

 Field name   | Mandatory? | Allowed values  | Allowed special characters 
--------------|------------|-----------------|----------------------------
 Minutes      | Yes        | 0-59            | * / , -                    
 Hours        | Yes        | 0-23            | * / , -                    
 Day of month | Yes        | 1-31            | * / , - ?                  
 Month        | Yes        | 1-12 or JAN-DEC | * / , -                    
 Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?                  

Special Characters ¶

### Asterisk ( * )

The asterisk indicates that the cron expression will match for all values of the field; e.g., using an asterisk in the
5th field (month) would indicate every month.

### Slash ( / )

Slashes are used to describe increments of ranges. For example 3-59/15 in the 1st field (minutes) would indicate the 3rd
minute of the hour and every 15 minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...", that
is, an increment over the largest possible range of the field. The form "N/..." is accepted as meaning "N-MAX/...", that
is, starting at N, use the increment until the end of that specific range. It does not wrap around.

### Comma ( , )

Commas are used to separate items of a list. For example, using "MON,WED,FRI" in the 5th field (day of week) would mean
Mondays, Wednesdays and Fridays.

### Hyphen ( - )

Hyphens are used to define ranges. For example, 9-17 would indicate every hour between 9am and 5pm inclusive.

### Question mark ( ? )

Question mark may be used instead of '*' for leaving either day-of-month or day-of-week blank.

## 設定ファイルでのCron管理

`config/config.<env>.yml` の `jobSchedulerConfig.cron` にジョブ名とCron式を定義すると、サーバー起動時に設定値がそのまま利用されます。  
キーには `RefreshTokenJob` や `CreateHolidayCalendarJob` など `server.go` 内で登録しているジョブ名を指定してください。
