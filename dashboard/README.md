# AIExpense Metrics Dashboard

A modern, real-time metrics dashboard for AIExpense built with Next.js, React, TypeScript, and shadcn/ui.

## Features

- 📊 Real-time metrics visualization
- 📈 Daily Active Users (DAU) tracking with line charts
- 💰 Expense aggregation with bar charts
- 👥 User growth metrics and analytics
- 🎨 Modern dark theme with Tailwind CSS
- 🔐 API key authentication
- 📱 Responsive design (mobile, tablet, desktop)

## Tech Stack

- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui + Radix UI
- **Charts**: Recharts
- **HTTP Client**: Axios
- **Runtime**: Bun (optional, works with Node.js too)

## Prerequisites

- Node.js 18+ or Bun 1.0+
- AIExpense backend running on `http://localhost:8080`
- Valid admin API key for metrics endpoints

## Installation

### With Bun

```bash
cd dashboard
bun install
```

### With npm/yarn

```bash
cd dashboard
npm install
# or
yarn install
```

## Configuration

1. Copy the environment template:
```bash
cp .env.example .env.local
```

2. Update `.env.local` with your API configuration:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Running the Dashboard

### Development

```bash
# With Bun
bun run dev

# With npm
npm run dev

# With yarn
yarn dev
```

The dashboard will be available at `http://localhost:3000`

### Production Build

```bash
# With Bun
bun run build
bun start

# With npm
npm run build
npm start
```

## Using the Dashboard

1. Open http://localhost:3000 in your browser
2. Enter your admin API key (default: `admin_key`)
3. View real-time metrics:
   - **Metrics Grid**: Key performance indicators
   - **Charts**: DAU and expense trends
   - **Growth Analytics**: User and expense growth rates

### API Key Management

- API keys are stored in browser localStorage for convenience
- To change API keys, click "Logout" and enter a new key
- For production, consider implementing secure token storage

## Project Structure

```
dashboard/
├── src/
│   ├── app/
│   │   ├── layout.tsx        # Root layout
│   │   ├── page.tsx          # Dashboard page
│   │   └── globals.css       # Global styles
│   ├── components/
│   │   ├── Header.tsx        # Dashboard header
│   │   ├── MetricsGrid.tsx   # Key metrics cards
│   │   └── ChartSection.tsx  # Chart visualizations
│   └── lib/
│       └── api.ts            # API utilities (optional)
├── public/                   # Static assets
├── package.json
├── tsconfig.json
├── next.config.js
├── tailwind.config.ts
├── postcss.config.js
└── README.md
```

## API Integration

The dashboard connects to the following AIExpense API endpoints:

```
GET /api/metrics/dau
  - Returns daily active users over time
  - Requires X-API-Key header
  - Response: { data: DailyMetrics[], total_active_users, average_daily_users }

GET /api/metrics/expenses-summary
  - Returns daily expense aggregates
  - Requires X-API-Key header
  - Response: { data: DailyMetrics[], total_expenses, average_daily_expenses, ... }

GET /api/metrics/growth
  - Returns system growth metrics
  - Requires X-API-Key header
  - Response: { total_users, new_users_today, new_users_this_week, new_users_this_month, ... }
```

## Customization

### Theme Colors

Edit `src/app/globals.css` and `tailwind.config.ts` to customize:
- Primary colors
- Background colors
- Accent colors
- Font sizes

### Charts

Edit `src/components/ChartSection.tsx` to:
- Add new chart types
- Customize chart colors
- Change data visualization

### Metrics Cards

Edit `src/components/MetricsGrid.tsx` to:
- Add custom metrics
- Modify card layout
- Add trend indicators

## Troubleshooting

### Dashboard not loading metrics

1. Verify backend is running on `http://localhost:8080`
2. Check API key is correct
3. Ensure CORS is enabled on backend
4. Check browser console for errors

### Build errors

```bash
# Clear Next.js cache
rm -rf .next

# Reinstall dependencies
bun install  # or npm install

# Rebuild
bun run build  # or npm run build
```

### Slow performance

1. Check browser network tab for slow API calls
2. Verify backend is responsive
3. Clear browser cache and reload
4. Check system resources (CPU, memory)

## Deployment

### Docker

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]
```

```bash
docker build -t aiexpense-dashboard .
docker run -p 3000:3000 -e NEXT_PUBLIC_API_URL=http://backend:8080 aiexpense-dashboard
```

### Vercel

```bash
# Install Vercel CLI
npm install -g vercel

# Deploy
vercel
```

Set environment variables in Vercel dashboard:
```
NEXT_PUBLIC_API_URL=https://api.your-domain.com
```

### Traditional Server

```bash
# Build
npm run build

# Start server (ensure PORT=3000)
npm start
```

Or use PM2:
```bash
npm install -g pm2
pm2 start npm --name "dashboard" -- start
pm2 save
```

## Contributing

Contributions are welcome! Please:
1. Create a feature branch
2. Make your changes
3. Test thoroughly
4. Submit a pull request

## License

MIT - Same as AIExpense project

## Support

For issues or questions:
1. Check the troubleshooting section
2. Review backend logs
3. Check browser console for errors
4. Open an issue on GitHub
