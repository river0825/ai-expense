# Change: Add Webchat Interface

## Why
To easily test the terminal chat API from the browser without needing curl or other tools. This provides a visual interface for developers and users to interact with the expense tracking bot.

## What Changes
- Add a new `/chat` page in the Next.js dashboard
- Create a chat UI component with message input and message history
- Connect to the existing `/api/chat/terminal` backend endpoint
- Store user_id in localStorage for session persistence

## Impact
- **Specs**: `dashboard-webchat` (Added)
- **Code**:
  - `dashboard/src/app/chat/page.tsx` - Chat page
  - `dashboard/src/components/ChatInterface.tsx` - Chat UI component
