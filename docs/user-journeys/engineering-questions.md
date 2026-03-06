# Engineering Questions（資深工程師待確認）

> 這份文件收錄「過於工程細節」或「目前 PO 不宜直接承諾」的問題。

## A. 安全與連結存取

1. short link 是否一次性？可否重放？
2. token 在 URL query + cookie 同時存在時，安全風險與治理策略為何？
3. `report_token` cookie 的 `HttpOnly` 為何設為 false，安全補償措施是什麼？

## B. 金流資料正確性

4. default input currency 與 home currency 的資料一致性檢查在哪一層保證？
5. 匯率來源、更新頻率、失敗 fallback、歷史一致性策略是什麼？
6. 多幣別輸入混合在同一訊息（例如 JPY + USD）目前是否支援？

## C. 路由與意圖

7. Regex intent 與 AI intent 衝突時，最終優先序是否固定且有回歸測試？
8. AI intent `UNKNOWN` 的觀測指標與門檻是什麼（何時要調 prompt）？
9. 旅行情境判斷的 confidence 門檻是否可配置？

## D. 可靠性與資料一致性

10. 多筆記帳中若某筆 create 失敗，現在是部分成功；是否需要交易化（全成全敗）選項？
11. idempotency key 目前依賴 message ID；跨平台重送或 message ID 缺失時如何處理？
12. account/wallet 名稱正規化字典與治理流程在哪裡（避免同義詞分裂）？

## E. 測試與觀測

13. user journey -> test case 的 coverage matrix 是否可自動生成？
14. 目前 journey 成功率、失敗率、延遲是否已有 dashboard？
15. 關鍵失敗（AI timeout、匯率失敗、短連結生成失敗）是否有 alert？

## F. 多平台一致性

16. LINE/Slack/Telegram/Teams/Discord/WhatsApp/Terminal 的能力差異表是否存在？
17. 哪些 journey 是 platform-specific，哪些 truly shared？

## 已完成確認

- Report link 文案與實作有效期已對齊為 7 天（原先文案 5 分鐘不一致已修正）。
