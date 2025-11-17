import { createBrowserRouter } from 'react-router-dom';

import { App } from '@/components/layout/App';

import { RoutePaths } from '@/shared/config/Routes';

import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { DashboardPage } from './pages/DashboardPage';
import { LoginPage } from './pages/LoginPage';
import { SignupPage } from './pages/SignupPage';

const router = createBrowserRouter([
  {
    path: RoutePaths.home,
    element: <App />,
    children: [
      {
        index: true,
        element: (
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        ),
      },
      {
        path: RoutePaths.signup,
        element: (
          <SignupPage />
        ),
      },
      {
        path: RoutePaths.login,
        element: (
          <LoginPage />
        ),
      },
    ],
  },
]);

export default router;