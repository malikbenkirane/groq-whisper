import 'package:go_router/go_router.dart';
import 'package:theatre/routing/routes.dart';
import 'package:theatre/ui/screens/home_screen.dart';

GoRouter router() => GoRouter(
  initialLocation: Routes.home,
  debugLogDiagnostics: true,
  routes: [
    GoRoute(
      path: Routes.home,
      builder: (context, state) {
        return HomeScreen();
      },
    ),
  ],
);
