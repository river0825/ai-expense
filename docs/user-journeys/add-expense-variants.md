# 使用者旅程：Add Expense 與其變型（PO 視角）

狀態：`【已開發】`（主流程與多數變型）

## 這份文件是做什麼的

這份文件把「Add Expense」當作同一條主旅程，整理目前系統支援的不同變型，避免把相近流程分散成很多看不出關聯的文件。

---

## Add Expense 主旅程（共通骨幹）

1. 使用者輸入一段記帳訊息（自然語言）
2. 系統解析出可用欄位（例如金額、描述、幣別、日期、帳戶）
3. 系統建立一筆或多筆 expense
4. 系統回傳確認訊息（含摘要）

---

## 變型清單（Variants）

### Variant A：單筆記帳
- 狀態：`【已開發】`
- 例子：`早餐 80`
- 預期：建立 1 筆 expense 並回覆成功
- 參考：`internal/usecase/process_message_test.go`

#### 使用者操作（示例）

```text
早餐 80
```

#### 系統回應（示例）

```text
✓ Recorded 1 expense(s), total: 80 TWD
```

### Variant B：多筆記帳（同一則訊息）
- 狀態：`【已開發】`
- 例子：`早餐$20午餐$30晚餐$50`
- 預期：拆分後建立多筆 expense
- 參考：`test/e2e/webhook_flow_test.go`

#### 使用者操作（示例）

```text
早餐$20午餐$30晚餐$50
```

#### 系統行為（摘要）

- 將單一訊息拆成多筆 parsed expense
- 逐筆建立 expense
- 回覆彙總結果

### Variant C：含幣別輸入
- 狀態：`【已開發】`
- 例子：`Lunch 120 USD with Visa`
- 預期：保留原始幣別並換算本位幣
- 參考：`internal/adapter/http/api_journey_test.go`

### Variant D：含帳戶/付款方式輸入（wallet/account）
- 狀態：`【已開發】`
- 例子：`午餐 200 用玉山卡`（或英文等價表達）
- 預期：expense 寫入 account 欄位，後續可於 dashboard 篩選/統計
- 參考：`internal/domain/aggregate.go`, `internal/usecase/create_expense.go`, `frontend/dashboard/src/components/AccountFilter.tsx`

### Variant E：旅行後的預設幣別記帳
- 狀態：`【已開發】`
- 例子：先確認切到 JPY，再輸入 `ramen 500`
- 預期：500 以 JPY 解讀，系統回傳本位幣換算結果
- 參考：`docs/user-journeys/currency-journeys.md`

### Variant F：重複訊息去重（idempotency）
- 狀態：`【已開發】`
- 例子：同一 message ID 重送
- 預期：避免重複寫入 expense
- 參考：`internal/usecase/process_message_test.go`

---

## PO 想像（尚未開發）

### Variant G：收據圖片直接入帳
- 狀態：`【尚未開發】`
- 想像：使用者上傳收據後，自動抽取金額/日期/商家並建議分類，使用者只需確認。

### Variant H：不確定欄位的互動式補問
- 狀態：`【尚未開發】`
- 想像：若只缺少少數欄位（例如帳戶），系統只補問必要問題，而不是整筆重來。

### Variant I：批次修正
- 狀態：`【尚未開發】`
- 想像：多筆記帳後，使用者可用一句話修正其中某筆（例如「第二筆改 35」）。

---

## 對 PM/PO 的價值

- 用一張表看懂 Add Expense 的能力邊界
- 可以快速標示哪些是「已可對外承諾」與哪些是「產品想像」
- 方便拆成 roadmap（先補問，再收據，再批次修正）

## 相關文件

- `docs/user-journeys/index.md`
