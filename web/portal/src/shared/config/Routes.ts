export const Routes = [
  { key: 'landing', label: 'Landing Page', path: '/', end: true },
  { key: 'dashboard', label: 'Dashboard', path: '/dashboard', end: true },
]

export type RouteKey = (typeof Routes)[number]['key']

export const RoutePaths = Routes.reduce((acc, route) => {
  acc[route.key] = route.path
  return acc
}, {} as Record<RouteKey, string>)