export default function GlobalNotFound() {
  return (
    <html lang="en">
      <body>
        <div style={{ 
          display: 'flex', 
          flexDirection: 'column',
          alignItems: 'center', 
          justifyContent: 'center', 
          minHeight: '100vh',
          fontFamily: 'system-ui, sans-serif',
          backgroundColor: '#0f172a',
          color: '#f1f5f9'
        }}>
          <h1 style={{ fontSize: '4rem', margin: 0 }}>404</h1>
          <p style={{ color: '#94a3b8' }}>Page not found</p>
          <a 
            href="/dashboard" 
            style={{ 
              marginTop: '1rem',
              padding: '0.5rem 1rem',
              backgroundColor: '#f59e0b',
              color: '#0f172a',
              borderRadius: '0.5rem',
              textDecoration: 'none',
              fontWeight: 500
            }}
          >
            Go to Dashboard
          </a>
        </div>
      </body>
    </html>
  );
}
