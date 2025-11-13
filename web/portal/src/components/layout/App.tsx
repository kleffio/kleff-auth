import { Outlet } from 'react-router-dom';

export function App() {
  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col">
      <main className="app-container flex-1 py-10">
        <Outlet />
      </main>
    </div>
  );
}