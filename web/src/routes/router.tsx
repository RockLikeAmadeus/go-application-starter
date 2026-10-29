import { createBrowserRouter, type RouteObject } from "react-router";
import { ItemsPage } from "../features/items/ItemsPage";
import { Layout, NotFound } from "./Layout";

export const routes: RouteObject[] = [
  {
    element: <Layout />,
    children: [
      { index: true, element: <ItemsPage /> },
      { path: "*", element: <NotFound /> },
    ],
  },
];

export const router = createBrowserRouter(routes);
