# 2024 Dcard Backend Assignments

> [!NOTE]
> ### Two implementations
> 1. Encore 框架 (放在[encore branch](https://github.com/TTC-CCF/Dcard-Backend-Assignment/tree/encore))
> 2. Gin 框架  
>
> 一開始我是先用 Encore 框架實作，但是發現Encore因為沒辦法把內建的一些middleware關掉，導致一直沒有辦法達到10000qps，所以後來改用Gin框架實作。  
> 兩種實作方式是差不多的，資料庫都是用postgresql，並且都有使用gorm這個ORM框架。

### Overview
使用Gin框架實作廣告投放API，主要的路由有:
- `GET /api/v1/ad` 投放廣告
- `POST /api/v1/ad` 建立廣告  
以[API spec](https://drive.google.com/file/d/1dnDiBDen7FrzOAJdKZMDJg479IC77_zT/view?usp=sharing)為主的API設計。

### Project Structure
以MVC架構設計:
```bash
src/
├── cache/
│   ├── cache.go
├── controllers/
│   ├── banner_controller.go
├── models/
│   ├── banner_model.go
├── ├── connections.go
├── routes/
│   ├── banner_router.go
├── tests/
│   ├── api_test/
│   │   ├── api_test.go
│   ├── load_test/
│   │   ├── script.js  # k6 load test
│   ├── unit_test/
│   │   ├── cache_test.go
├── utils/
│   ├── params.go   # 各種結構
├── main.go
```

### Setup
- **Prerequisite**  
    Go 1.13以上
    Docker, GNU Make, K6
- **Run**  
    1. Git clone  
        ```bash
        git clone https://github.com/TTC-CCF/Dcard-Backend-Assignment.git
        cd Dcard-Backend-Assignment
        ```  
    2. 設定.env (資料庫、redis連線、PORT等等)  
        ```bash
        cp .env.example .env
        ```
    3. Run  
        ```bash
        make
        ```
    預設執行api server在localhost:3000
    

### Database Schema
因為一個廣告可以有多種條件，而且每個條件都是一個固定的集合，基於節省資料儲存空間和優化資料庫查詢效率，我設計many to many的結構。  
此設計達到3NF，如下圖所示:
![](/assets/er_diagram.png)

另外，因為gorm會在many to many的conjuction table上面建立index，所以資料庫查詢效率會更好。

### Cache Data
使用`go-redis`套件，將廣告條件的結果快取起來，降低資料庫負擔也提高QPS。
我用了兩組快取來維護資料一致性:
1. key: url path, value: response data
2. key: (age | gender | country | platform), value: url path
第一種快取是為了快速回傳結果，第二種是在建立新廣告的時候，刪除對應條件的快取。
以下面的request為例:
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
這個請求會刪除`age`快取裡面的所有key的第一種快取，並刪除`age`快取

### Testing
```bash
cd src/tests
make testAll    # run all tests
make unitTest   # run unit test
make apiTest    # run api test
make loadTest   # run load test
```  
- Unit Test: 
    - 針對`cache.go`裡的一些function做單元測試
    - 使用[redismock](https://github.com/go-redis/redismock)
- API Test:
    - 測試API參數驗證
    - 測試API回傳結果
    - 在testDB裡面測試
- Load Test:
    - 使用[k6](https://k6.io/)壓力測試
    - 模擬隨機url query
    - 測試結果:  
        12th Gen Intel(R) Core(TM) i7-12700H，2700 Mhz，14 Cores，20 Logical Processor  
        21176 requests/second
        ![](/assets/load_test.png)
