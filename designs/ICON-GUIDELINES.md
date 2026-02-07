# AIExpense App Icon Guidelines

**設計者：** Claude Code + ui-ux-pro-max
**建立日期：** 2026-02-07
**風格：** Minimalist Monogram (Modern + AI-Native)

---

## 核心概念

AIExpense icon 採用 **Minimalist Monogram** 設計，將三個核心要素完美結合：

1. **品牌標識** - 字母 "A" 代表 AIExpense
2. **聊天對話** - 氣泡和橫線代表對話式記帳
3. **發送動作** - 橙色箭頭代表即時提交和轉帳

## 色彩系統

### 主色系
| 用途 | 顏色 | Hex Code | RGB | 說明 |
|------|------|----------|-----|------|
| 背景漸層 1 | 藍色 | #2563EB | 37, 99, 235 | 左上角 |
| 背景漸層 2 | 橙色 | #F97316 | 249, 115, 22 | 右下角 |
| 主要線條 | 深藍 | #2563EB | 37, 99, 235 | 字母 A 和聊天氣泡 |
| 強調色 | 橙色 | #F97316 | 249, 115, 22 | 發送箭頭 |
| 背景 | 白色 | #FFFFFF | 255, 255, 255 | 中心背景 |

### 漸層使用
```css
/* CSS Gradient */
background: linear-gradient(135deg, #2563EB 0%, #F97316 100%);

/* SVG Gradient */
<linearGradient id="bgGradient" x1="0%" y1="0%" x2="100%" y2="100%">
  <stop offset="0%" style="stop-color:#2563EB;stop-opacity:1" />
  <stop offset="100%" style="stop-color:#F97316;stop-opacity:1" />
</linearGradient>
```

## 圓角規則

| 應用場景 | 圓角半徑 |
|----------|--------|
| iOS App Icon | 圓形（對應 iOS 14+） |
| Android App Icon | 圓形 |
| Web Favicon | 0-8px（視情況） |
| 應用商店 | 圓形 / 圓角正方形 |

## 各平台尺寸規格

### iOS 應用
| 用途 | 尺寸 | 設備 | 倍率 |
|------|-----|------|------|
| App Icon | 192x192 | iPhone 6-8, X, 11 | @3x (1x = 64x64) |
| App Icon | 180x180 | iPhone 6-8 Plus | @3x (1x = 60x60) |
| App Icon | 120x120 | iPhone 6s, 7, 8 | @2x (1x = 60x60) |
| Spotlight | 120x120 | iPhone | @2x (1x = 60x60) |
| Settings | 87x87 | iPhone | @3x (1x = 29x29) |

**iOS Icon 要求：**
- 圓形或完全填充正方形
- 無圓角（系統自動應用 mask）
- 完整覆蓋安全區域（無邊距）
- 導出格式：PNG 32-bit (RGBA)

### Android 應用
| 用途 | 尺寸 | DPI | 說明 |
|------|-----|-----|------|
| App Icon | 192x192 | xxhdpi | 標準應用圖標 |
| App Icon | 144x144 | xhdpi |
| App Icon | 96x96 | hdpi |
| App Icon | 72x72 | mdpi |

**Android Icon 要求：**
- 正方形（108x108dp minimum）
- 圖標需預留 8dp 內邊距
- 導出格式：PNG 32-bit (RGBA)
- 檔案路徑：`res/mipmap-{dpi}/ic_launcher.png`

### Web / Favicon
| 用途 | 尺寸 | 格式 | 說明 |
|------|-----|------|------|
| Favicon | 32x32 | ICO / PNG | 瀏覽器標籤 |
| Apple Touch Icon | 180x180 | PNG | iOS 首屏 |
| Android Chrome | 192x192 | PNG | Chrome Android |
| Web App Icon | 512x512 | PNG | PWA manifest |

**Web Icon 要求：**
- 導出格式：PNG 或 SVG
- 檔案大小：< 50KB (PNG), < 20KB (SVG)
- 色彩空間：sRGB

### 應用商店
| 平台 | 尺寸 | 格式 | 說明 |
|------|-----|------|------|
| App Store | 1024x1024 | PNG/JPG | App Store 展示 |
| Google Play | 512x512 | PNG | Google Play 展示 |
| 微信 | 120x120 | PNG | 微信小程序 |
| Telegram Bot | 160x160 | PNG/SVG | Telegram 機器人 |

## 設計約束條件

### 必須遵守
- ✅ 使用指定的色彩系統
- ✅ 保持 "A" 字母清晰可辨
- ✅ 聊天氣泡和箭頭保持可見
- ✅ 最小尺寸 32x32px 時仍保持識別度
- ✅ 無圖文混合（icon 獨立）

### 禁止項目
- ❌ 改變色彩（除了深色模式調整）
- ❌ 添加投影或複雜效果
- ❌ 旋轉或扭曲
- ❌ 添加文字（APP 名稱）
- ❌ 簡化細節（即使在小尺寸）

## 深色模式適配

### 深色背景應用
若在深色背景（如 Android 深色主題）上使用，建議：

```css
/* 深色模式版本 */
background: linear-gradient(135deg, #1E3A8A 0%, #EA580C 100%);
/* 或使用反色 */
background: white;
color: linear-gradient(135deg, #2563EB 0%, #F97316 100%);
```

## 導出清單

### SVG 版本
```
icon-aiexpense.svg (可縮放，用於網頁和設計工具)
```

### PNG 版本（推薦導出）
```
192x192/icon-aiexpense.png (推薦起點)
120x120/icon-aiexpense.png (iOS Spotlight)
512x512/icon-aiexpense.png (應用商店)
```

### 設計檔案
```
icon-aiexpense.sketch (Sketch)
icon-aiexpense.figma (Figma)
icon-aiexpense.psd (Photoshop)
```

## 實現步驟

### 1. 設計工具中使用
1. 在 Figma / Sketch 中打開 SVG 或設計檔案
2. 選擇合適尺寸的 artboard
3. 導出為 PNG (32-bit RGBA)
4. 壓縮使用線上工具 (TinyPNG)

### 2. iOS 應用
```bash
# 使用 ImageMagick 生成多個尺寸
convert icon-192.png -resize 180x180 ios/iPhone-6-Plus.png
convert icon-192.png -resize 120x120 ios/iPhone-spotlight.png
convert icon-192.png -resize 87x87 ios/iPhone-settings.png
```

### 3. Android 應用
```bash
# 生成 Android 資源
mkdir -p android/mipmap-{mdpi,hdpi,xhdpi,xxhdpi}
convert icon-192.png -resize 72x72 android/mipmap-mdpi/ic_launcher.png
convert icon-192.png -resize 96x96 android/mipmap-hdpi/ic_launcher.png
convert icon-192.png -resize 144x144 android/mipmap-xhdpi/ic_launcher.png
convert icon-192.png -resize 192x192 android/mipmap-xxhdpi/ic_launcher.png
```

### 4. Web 應用
```html
<!-- Favicon -->
<link rel="icon" type="image/png" href="/favicon-32x32.png" sizes="32x32" />
<link rel="icon" type="image/png" href="/favicon-192x192.png" sizes="192x192" />

<!-- Apple Touch Icon -->
<link rel="apple-touch-icon" href="/apple-touch-icon.png" />

<!-- Android Chrome -->
<link rel="icon" type="image/png" href="/android-chrome-192x192.png" sizes="192x192" />

<!-- PWA Manifest -->
{
  "icons": [
    {
      "src": "/android-chrome-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/android-chrome-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

## 品牌一致性

AIExpense icon 應與以下材料保持一致：

- ✅ Landing Page 色彩系統（藍色 #2563EB + 橙色 #F97316）
- ✅ Logo 和品牌指南
- ✅ 應用 UI 主色調
- ✅ 行銷材料配色

## 常見問題

### Q: 可以改變 icon 的顏色嗎？
A: 不建議。指定色彩是品牌識別的一部分。若有特殊需求，請諮詢設計團隊。

### Q: 最小尺寸是多少？
A: 建議不小於 32x32px。小於此尺寸細節會丟失。

### Q: SVG 版本可用於列印嗎？
A: 可以，但建議先轉換為 PNG。SVG 在某些列印軟體中可能有相容性問題。

### Q: 如何使用 icon 在網站上？
```html
<!-- SVG 直接嵌入 -->
<svg width="64" height="64" viewBox="0 0 512 512">...</svg>

<!-- 或使用圖片 -->
<img src="icon-aiexpense.png" alt="AIExpense" width="64" height="64" />
```

### Q: 深色模式下 icon 顯示不清楚怎麼辦？
A: 使用深色背景版本或添加白色背景圓形來提高對比度。

## 版本歷史

| 版本 | 日期 | 變更 |
|------|------|------|
| 1.0 | 2026-02-07 | 初始發佈 - Minimalist Monogram 設計 |

---

**最後更新：** 2026-02-07
**設計規格版本：** v1.0
**狀態：** ✅ 可用於生產
