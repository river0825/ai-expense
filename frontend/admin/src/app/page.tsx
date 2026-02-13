export default function AdminPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <h1 className="text-4xl font-bold mb-4">Admin Panel</h1>
      <p className="text-xl text-gray-400 mb-8">Welcome to the AIExpense Admin Panel</p>
      <a href="/dashboard" className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors">
        Go to Dashboard
      </a>
    </main>
  );
}
