# 2024 Dcard backend assignment
### Overview
使用Golang [Encore](https://encore.dev/)後端框架開發，實作一個簡易的廣告投放服務。
包含以下API：
- **Admin API**:
    - 新增廣告
- **Public API**:
    - 投放廣告  

>[!NOTE]
>此兩個API並沒有驗證機制，只是簡易的示範API*  
>[API Spec](https://drive.google.com/file/d/1dnDiBDen7FrzOAJdKZMDJg479IC77_zT/view?usp=sharing)

高併發設計用到的技術：
- **Cache:**  
    使用Encore內建的快取套件，將廣告資料快取在記憶體中，降低資料庫負擔以及提升讀取速度。我所使用的Cache Strategy是LRU，當快取空間不足時，會將最少被使用的資料移除。在更新資料時也會刪除快取內的資料，保持資料一致性。
- **PostgreSQL GIN Index:**  
    在常被query的欄位上建立GIN Index，提升查詢速度。因為我有使用到Postgre的ARRAY資料型態來儲存`country`、`gender`以及`platform`，而且Public API會根據傳入參數的不同，組合不同的查詢條件，所以用GIN Index來提升查詢速度。建立Index的語法可見[migration檔案](https://github.com/TTC-CCF/Dcard-Backend-Assignment/tree/master/v1/migrations)。
- **SingleFlight**
    因為我有使用到Cache來降低資料庫負擔，而Public API又需要高併發的能力，所以我使用SingleFlight來避免Cache穿透的問題。當多個goroutine同時存取Cache時，會將相同的key合併成一個request，只有一個goroutine會去存取資料庫。

### Project Structure
`v1`: 
- `migration`: 資料庫遷移檔案
- `ad.go`: API邏輯，包含資料庫操作
- `ad_test.go`: API測試
### How to run
1. [安裝Encore CLI](https://encore.dev/docs/quick-start), Go 1.16以上
2. `git clone`後，`cd`到專案目錄
3. `encore run` 

>[!NOTE]
>在encore主控介面可以呼叫API，但是GET的功能好像沒辦法用他的介面傳遞url query參數，所以建議使用Postman或其他工具呼叫API