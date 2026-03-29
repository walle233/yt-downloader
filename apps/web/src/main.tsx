import { ClerkProvider } from "@clerk/react";
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { initAnalytics } from "./lib/analytics";
import "./styles.css";

const clerkPublishableKey = window.__RUNTIME_CONFIG__?.VITE_CLERK_PUBLISHABLE_KEY || import.meta.env.VITE_CLERK_PUBLISHABLE_KEY || "";
const root = ReactDOM.createRoot(document.getElementById("root")!);

initAnalytics();

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
      <main className="mx-auto flex min-h-screen w-full max-w-4xl items-center px-4 py-12">
        <div className="w-full rounded-[2rem] border border-[#ffdad6] bg-white p-10 shadow-[0_18px_44px_rgba(28,27,27,0.07)]">
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-[#956d67]">Archive</p>
          <h1 className="mt-3 font-display text-4xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">Clerk is not configured.</h1>
          <p className="mt-4 text-lg leading-8 text-[#603e39]">
            Set <code>VITE_CLERK_PUBLISHABLE_KEY</code> in runtime config before using the product UI.
          </p>
        </div>
      </main>
    </React.StrictMode>,
  );
}
