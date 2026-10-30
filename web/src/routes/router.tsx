import { createBrowserRouter, type RouteObject } from "react-router";
// sample:start
import { ItemsPage } from "../features/items/ItemsPage";
// sample:end
import { Home, Layout, NotFound } from "./Layout";

export const routes: RouteObject[] = [
  {
    element: <Layout />,
    children: [
      // The first index route wins, so Home only shows once the sample is removed.
      // sample:start
      { index: true, element: <ItemsPage /> },
      // sample:end
      { index: true, element: <Home /> },
      { path: "*", element: <NotFound /> },
    ],
  },
];

export const router = createBrowserRouter(routes);
