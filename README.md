# GAE Search API サンプル

[公式リファレンス](https://cloud.google.com/appengine/docs/standard/go/search/)
[GoDoc](https://godoc.org/google.golang.org/appengine/search)


## Datastore
### User
* Key  
  UUID

| プロパティ | 説明           | データ型     |
| ---        | ---            | ---          |
| Name       | ユーザ名       | 文字列       |
| Comment    | コメント       | 文字列       |
| Visits     | 訪問数         | 小数         |
| LastVisit  | 前回訪問日時   | 日時         |
| Birthday   | 誕生日         | 日時         |
| Mail       | メールアドレス | 文字列リスト |
| UserID     | ユーザID       | 整数値       |


## Search
### UserIndex
* ID  
  Datastore#User#Key

| プロパティ | 説明                                               | データ型 |
| ---        | ---                                                | ---      |
| Name       | ユーザ名                                           | 文字列   |
| Comment    | コメント                                           | HTML     |
| Visits     | 訪問数                                             | 小数     |
| LastVisit  | 前回訪問日時                                       | 日時     |
| Birthday   | 誕生日                                             | 日時     |
| Mail       | メールアドレス<br>複数存在する場合はスペース区切り | 文字列   |
| UserID     | ユーザID                                           | 整数値   |


## メモ
* RDBのLIKE検索のようなものはない？
  - 試した感じ、完全一致ともライク検索とも言えない感じ
  - どんな感じで検索条件にマッチするのかちょっと調べたほうが良さそう
