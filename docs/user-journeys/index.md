# User Journey Index

這份文件整理目前系統已存在的使用者旅程（user journeys），聚焦在「使用者怎麼使用系統」，不是規格書。

> 文件標註：`【已開發】` 代表目前可用；`【尚未開發】` 代表 PO 想像中的目標體驗。

## 目前優先維護（Top 5）

- `add-expense-variants.md`
- `currency-journeys.md`
- `first-time-auto-signup.md`
- `report-link-to-dashboard.md`

## 決策與問答

- PO 可直接回答：`docs/user-journeys/po-qa.md`
- 工程細節待確認：`docs/user-journeys/engineering-questions.md`
- 工程整理待辦：`docs/user-journeys/engineering-backlog.md`

## Onboarding 與啟用

### 1. 第一次使用自動註冊
- 狀態：`【已開發】`
- 入口：使用者第一次透過聊天介面發訊息
- 使用者動作：直接輸入一般訊息
- 系統行為：自動建立使用者與預設分類，接著處理該訊息
- 參考：`internal/usecase/gherkin_scenarios_test.go`, `test/e2e/webhook_flow_test.go`

### 2. 既有使用者再次互動
- 狀態：`【已開發】`
- 入口：既有使用者再次發訊
- 使用者動作：送出新訊息
- 系統行為：不重複建立帳號，直接進入處理流程
- 參考：`internal/usecase/gherkin_scenarios_test.go`

### 3. 多平台進入
- 狀態：`【已開發】`
- 入口：LINE / Telegram / Slack / Teams / Discord / WhatsApp / Terminal
- 使用者動作：透過不同 messenger 觸發訊息
- 系統行為：由各平台 adapter 接入，共用核心用例流程
- 參考：`internal/usecase/gherkin_scenarios_test.go`, `docs/plans/2026-02-25-user-journeys-diagrams.md`

## 核心記帳流程

### 4. 聊天記帳（單筆）
- 狀態：`【已開發】`
- 入口：聊天訊息
- 使用者動作：輸入一筆消費（自然語言）
- 系統行為：解析金額/描述/分類/帳戶/日期，建立 expense
- 參考：`internal/usecase/process_message_test.go`, `internal/usecase/gherkin_scenarios_test.go`
- Add Expense 變型總覽：`docs/user-journeys/add-expense-variants.md`

### 5. 聊天記帳（多筆）
- 狀態：`【已開發】`
- 入口：聊天訊息
- 使用者動作：一則訊息內輸入多筆消費
- 系統行為：拆分成多筆 expense 並逐筆入庫
- 參考：`test/e2e/webhook_flow_test.go`
- Add Expense 變型總覽：`docs/user-journeys/add-expense-variants.md`

### 6. 重複訊息去重（Idempotency）
- 狀態：`【已開發】`
- 入口：帶有相同 message ID 的重送訊息
- 使用者動作：同訊息重送
- 系統行為：避免重複建立 expense
- 參考：`internal/usecase/process_message_test.go`

## 幣別與意圖流程

### 7. 幣別切換（關鍵字流程）
- 狀態：`【已開發】`
- 入口：聊天訊息
- 使用者動作：輸入切換預設幣別意圖
- 系統行為：必要時先澄清，再更新 default input currency
- 參考：`internal/usecase/process_message_test.go`
- 整合旅程文件：`docs/user-journeys/currency-journeys.md`

### 8. 旅行情境主動建議切換幣別（新）
- 狀態：`【已開發】`
- 入口：聊天訊息
- 使用者動作：描述旅行情境（例如在日本旅行）
- 系統行為：主動詢問是否切換幣別，確認後套用
- 參考：`internal/usecase/process_message_test.go`
- 整合旅程文件：`docs/user-journeys/currency-journeys.md`

### 9. AI Intent Fallback 路由
- 狀態：`【已開發】`
- 入口：關鍵字流程沒命中時
- 使用者動作：一般自然語言訊息
- 系統行為：AI 將訊息分流為 `TRAVEL_CONTEXT` / `CURRENCY_CHANGE` / `REPORT` / `ADD_EXPENSE` / `UNKNOWN`
- 參考：`internal/usecase/process_message.go`, `internal/usecase/process_message_test.go`

## 查詢、報表與 Dashboard

### 10. 查報表（Chat -> Magic Link）
- 狀態：`【已開發】`
- 入口：聊天訊息
- 使用者動作：輸入 report 類請求
- 系統行為：回傳短連結
- 參考：`internal/adapter/http/api_journey_test.go`, `README.md`

### 11. 短連結登入 Dashboard
- 狀態：`【已開發】`
- 入口：點擊短連結
- 使用者動作：開啟 report short link
- 系統行為：換發 token，進入 dashboard session
- 參考：`internal/adapter/http/api_journey_test.go`

### 12. Dashboard 編輯費用
- 狀態：`【已開發】`
- 入口：dashboard API
- 使用者動作：修改 expense
- 系統行為：更新資料並回傳成功
- 參考：`internal/adapter/http/api_journey_test.go`

### 13. 依日期區間查詢費用
- 狀態：`【已開發】`
- 入口：查詢流程
- 使用者動作：指定日期範圍
- 系統行為：回傳區間內費用
- 參考：`internal/usecase/gherkin_scenarios_test.go`

### 14. 刪除費用
- 狀態：`【已開發】`
- 入口：管理流程
- 使用者動作：刪除既有費用
- 系統行為：移除資料並反映到後續查詢
- 參考：`internal/usecase/gherkin_scenarios_test.go`

## 分類智慧與資料管理

### 15. AI 建議分類
- 狀態：`【已開發】`
- 入口：記帳解析流程
- 使用者動作：輸入消費描述
- 系統行為：推薦最可能分類
- 參考：`internal/usecase/gherkin_scenarios_test.go`

### 16. 分類修正學習
- 狀態：`【已開發】`
- 入口：分類管理流程
- 使用者動作：補充或修正分類關鍵字
- 系統行為：後續分類可依新關鍵字改善結果
- 參考：`internal/usecase/gherkin_scenarios_test.go`

---

## 已建立的旅程文件

- `docs/user-journeys/add-expense-variants.md`
- `docs/user-journeys/currency-journeys.md`
- `docs/user-journeys/first-time-auto-signup.md`
- `docs/user-journeys/report-link-to-dashboard.md`

## 後續建議

- 持續維持「主旅程 + 變型」架構，避免重複文件
- 每條旅程補上：
  - 使用者輸入範例
  - 系統回覆範例
  - 失敗/例外情境
  - 對應自動化測試檔案

## PO 想像地圖（尚未開發）

### A. 主動提醒與回訪
- 狀態：`【尚未開發】`
- 想像：若使用者一段時間沒有記帳，系統主動以合適語氣提醒並提供「一鍵快速記帳」入口。

### B. 預算超標即時預警
- 狀態：`【尚未開發】`
- 想像：當分類預算接近或超過門檻，系統在聊天中提供提醒與建議替代方案。

### C. 收據影像快速入帳
- 狀態：`【尚未開發】`
- 想像：使用者上傳收據後，自動抽取金額、日期、商家與分類，使用者只需確認。

### D. 報表對話式追問
- 狀態：`【尚未開發】`
- 想像：使用者問「這週花太多在哪？」系統可回覆重點並支援追問比較上週差異。
