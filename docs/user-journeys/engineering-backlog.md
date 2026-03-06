# Engineering Backlog（Journey 整理）

> 目的：把 `engineering-questions.md` 轉成可執行的工程工作。

## 已完成

1. **統一 report link 有效期文案與實作**
   - 優先級：P0
   - 狀態：Done
   - 說明：聊天回覆由「5 minutes」改為「7 days」，與 short link/token 實作一致。
   - 相關：`internal/usecase/process_message.go`

## 待辦（P0）

1. **定義 short link 一次性與重放策略**
   - 目標：明確「可否重放」「重放次數」「失效條件」
2. **確認 `report_token` cookie 安全策略**
   - 目標：釐清 `HttpOnly=false` 的必要性與補償控管
3. **補齊 idempotency 缺失場景**
   - 目標：message ID 缺失或跨平台重送時的防重策略

## 待辦（P1）

1. **幣別資料一致性規格化**
   - 目標：明確 default input currency vs home currency 的一致性檢查責任層
2. **匯率策略文件化**
   - 目標：來源、更新頻率、fallback、歷史一致性
3. **account/wallet 正規化字典**
   - 目標：避免同義詞碎片化（Cash/現金/Wallet 等）

## 待辦（P2）

1. **Journey -> Test Coverage Matrix 自動化**
2. **Journey Observability Dashboard**（成功率/失敗率/延遲）
3. **多平台能力差異表**（shared vs platform-specific）

## 建議 Owner 分工

- Backend/API：P0 #1, #2, #3
- Domain/Data：P1 #1, #3
- Infra/Platform：P1 #2, P2 #2
- QA/DevEx：P2 #1, #3
