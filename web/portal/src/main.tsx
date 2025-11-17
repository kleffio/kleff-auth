import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider } from 'react-router-dom';

import { ErrorBoundary } from "@/shared/errors/ErrorBoundary";
import { StatusView } from "@/shared/errors/components/StatusView";

import { AuthProvider } from "./features/auth/context/AuthContext";

import './index.css';

import router from './router';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary fallback={<StatusView status={503} message="Backend Offline" />}>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ErrorBoundary>
  </StrictMode>
);