# 使用者旅程：查報表（Chat -> Magic Link -> Dashboard）

狀態：`【已開發】`

## 這份文件是做什麼的

描述使用者從聊天請求報表，到打開短連結進入 dashboard 的完整流程。

---

## 角色與情境

- 角色：一般使用者
- 情境：想看自己的消費報表並進一步在 dashboard 操作

---

## 旅程步驟

### Step 1：使用者在聊天中要求報表

使用者輸入（示例）：

```text
show report
```

系統行為：

- 產生短連結（magic link）
- 回傳給使用者

### Step 2：使用者點擊短連結

系統行為：

- 透過短連結導轉
- 帶上有效 token，建立 dashboard 存取上下文

### Step 3：使用者進入 dashboard 查看與編輯

系統行為：

- 可查詢 expense
- 可更新 expense（例如修改描述與金額）

---

## 使用者價值

- 聊天入口與 dashboard 深度操作自然銜接
- 不需重新登入密碼流程（magic link）

---

## PO 對外說法（目前建議）

- 報表連結可讓使用者快速進入 dashboard 查詢與管理資料。
- 報表連結目前有效期為 7 天（短連結與 report token 同步）。

---

## PO 想像（未來版本）

- `【尚未開發】` 深連結上下文：使用者點開連結後可直接落在「本月摘要」或「特定分類」視圖。
- `【尚未開發】` 報表追問：在聊天中可直接追問「比上週多多少？」並帶入 dashboard 同步視圖。

---

## 參考與驗證

- `internal/adapter/http/api_journey_test.go`
