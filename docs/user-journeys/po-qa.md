# PO Q&A（依目前文件與現況）

> 角色說明：本文件由 PO 視角整理。技術不確定或過深細節，已轉到 `engineering-questions.md`。

## 已可明確回答

### 1) 目前最核心的使用者旅程是什麼？
- Add Expense 主旅程（含單筆、多筆、幣別、account/wallet 變型）
- 幣別旅程（關鍵字切換、旅行建議切換、輸入幣別與本位幣分離）
- First-time auto-signup
- Report link -> Dashboard

### 2) `【已開發】` 的定義是什麼？
- 在目前版本中已可使用，且在 code / tests 能找到對應證據。

### 3) Add Expense 現在支援哪些變型？
- 單筆記帳
- 多筆記帳（同一訊息）
- 幣別輸入與換算
- 帳戶/付款方式（account/wallet）
- 旅行後預設幣別記帳
- 重送訊息去重（idempotency）

### 4) 幣別相關現在可對外怎麼講？
- 可透過關鍵字切換預設輸入幣別
- 可在旅行情境下由系統主動建議切換（需使用者確認）
- 可把輸入幣別與本位幣分離：輸入可為 JPY/USD，報表維持本位幣比較

### 5) 報表旅程目前可對外怎麼講？
- 使用者可在聊天中請求 report
- 系統回傳短連結，使用者可進入 dashboard 查看與編輯資料
- 目前報表連結有效期為 7 天

## PO 目前不直接承諾（待工程確認）

- short link/token 的一次性/重放策略與安全邊界
- 多平台功能等價性矩陣（LINE/Slack/Telegram/Teams/Discord/WhatsApp/Terminal 是否完全一致）

以上項目已移交：`docs/user-journeys/engineering-questions.md`
