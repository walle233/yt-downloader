import { ClerkProvider } from "@clerk/react";
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./styles.css";

const clerkPublishableKey = window.__RUNTIME_CONFIG__?.VITE_CLERK_PUBLISHABLE_KEY || import.meta.env.VITE_CLERK_PUBLISHABLE_KEY || "";
const root = ReactDOM.createRoot(document.getElementById("root")!);

if (clerkPublishableKey) {
  root.render(
    <React.StrictMode>
      <ClerkProvider publishableKey={clerkPublishableKey}>
        <App authEnabled />
      </ClerkProvider>
    </React.StrictMode>,
  );
} else {
  root.render(
    <React.StrictMode>
      <App authEnabled={false} />
    </React.StrictMode>,
  );
}
