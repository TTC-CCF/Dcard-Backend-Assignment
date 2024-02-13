# 2024 Dcard backend assignment
### Overview
使用Golang [Encore](https://encore.dev/)後端框架開發，實作一個簡易的廣告投放服務。
包含以下API：
- **Admin API**:
    - 新增廣告
- **Public API**:
    - 投放廣告  

### API Doc
執行`encore run`後，可以在encore的主控介面看到API輸入和輸出，或是也可以直接參考[程式碼](https://github.com/TTC-CCF/Dcard-Backend-Assignment/blob/master/v1/ad.go)的api schema  
基本上跟[API Spec](https://drive.google.com/file/d/1dnDiBDen7FrzOAJdKZMDJg479IC77_zT/view?usp=sharing)相同

### Project Structure
`v1`: 
- `ad.go`: API邏輯
- `db.go`: 資料庫操作與Schema
- `ad_test.go`: API測試

### How to run
1. [安裝Encore CLI](https://encore.dev/docs/quick-start), Go 1.16以上
2. `git clone`後，`cd`到專案目錄
3. `encore run` 

### Techniques for achieving 10000 QPS
- **Cache:**  
    使用Encore內建的快取套件，將廣告資料快取在記憶體中，降低資料庫負擔以及提升讀取速度。快取的key為url query參數的組合，value為廣告資料。我也加了另外一個快取，分成`age`、`gender`、`country`、`platform`四個key，value為包含此條件的所有url query，新廣告加入後，會刪除新廣告擁有的條件的快取。例如  
    ```bash
    curl -X POST "http://localhost:4000/api/v1/ad"
     -H "Content-Type: application/json"
     -d '{
        "title": "test",
        "startAt": "2024-01-01T00:00:00Z",
        "endAt": "2024-12-31T23:59:59Z",
        "conditions": {
            "ageStart": 10,
            "ageEnd": 20,
        },
     }'
    ```
    這個請求會刪除key為`age`的快取，保持資料一致性。

- **SingleFlight:**  
    因為我有使用到Cache來降低資料庫負擔，而Public API又需要高併發的能力，所以我使用SingleFlight來避免Cache穿透的問題。當多個goroutine同時存取Cache時，會將相同的key合併成一個request，只有一個goroutine會去存取資料庫。

- **Database Normalization:**  
    對於Gender, Country, Platform三個欄位採用many to many的關係，加上GORM在conjunction table上會建立index，優化查詢效能。  
    此資料庫設計達到3NF
    ![er-diagram](/assets/er_diagram.png)

### Testing
使用Encore內建的測試框架(其實就是go test)，測試API邏輯以及資料庫操作。Encore測試的時候會額外建立一個測試用的docker db，然後測試之間用的cache也是獨立的，不會互相影響。
```
encore test ./...
```
`ad_test.go`:  
- `TestDeleteKeyspaceWhenCreate`  
    測試當新增廣告時，是否會刪除快取中的對應的key。

- `TestUpdateKeyspaceWhenRead`  
    測試當投放廣告時，是否會正確增加url query到快取對應的key中

- `TestAdmin`  
    測試Admin API，包含錯誤處理

- `TestPublic`  
    測試Public API，包含錯誤處理
