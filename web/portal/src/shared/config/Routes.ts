export const Routes = [
  { key: 'dashboard', label: 'Dashboard', path: '/', end: true },
  { key: 'login', label: 'Login', path: '/login', end: true },
  { key: 'signup', label: 'Signup', path: '/signup', end: true },
]

export type RouteKey = (typeof Routes)[number]['key']

export const RoutePaths = Routes.reduce((acc, route) => {
  acc[route.key] = route.path
  return acc
}, {} as Record<RouteKey, string>)