import { createBrowserRouter } from 'react-router-dom';

import { App } from '@/components/layout/App';

import { RoutePaths } from '@/shared/config/Routes';

import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { DashboardPage } from './pages/DashboardPage';
import { LandingPage } from './pages/LandingPage';

const router = createBrowserRouter([
  {
    path: RoutePaths.home,
    element: <App />,
    children: [
      {
        index: true,
        element: <LandingPage />,
      },
      {
        path: RoutePaths.dashboard,
        element: (
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        ),
      },
    ],
  },
]);

export default router;