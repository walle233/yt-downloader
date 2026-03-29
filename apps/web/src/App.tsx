import type { ReactNode } from "react";
import { BrowserRouter, Link, Route, Routes, useLocation } from "react-router-dom";
import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/react";
import { trackEvent } from "./lib/analytics";
import { DownloadPage } from "./pages/DownloadPage";
import { HomePage } from "./pages/HomePage";
import { SupportPage } from "./pages/SupportPage";
import { RouteAnalytics } from "./components/RouteAnalytics";

interface AppProps {
  authEnabled?: boolean;
}

function AppShell({ authEnabled, children }: { authEnabled: boolean; children: ReactNode }) {
  const location = useLocation();
  const onHome = location.pathname === "/";
  const onSupport = location.pathname === "/support";
  const onHistory = location.pathname.startsWith("/downloads/");

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(255,218,212,0.85),_transparent_26%),linear-gradient(180deg,_#fcf9f8_0%,_#f4efee_100%)] text-[#1c1b1b]">
      <header className="sticky top-0 z-50 border-b border-[#ebbbb4]/30 bg-[#fcf9f8]/85 backdrop-blur-xl">
        <div className="mx-auto flex w-full max-w-7xl items-center justify-between px-4 py-4 sm:px-6">
          <div className="flex items-center gap-3">
            <div className="grid size-11 place-items-center rounded-2xl bg-gradient-to-br from-[#bc0100] to-[#eb0000] text-white shadow-[0_16px_38px_rgba(188,1,0,0.22)]">
              <span className="material-symbols-outlined text-[22px]">slow_motion_video</span>
            </div>
            <div>
              <Link to="/" className="font-display text-2xl font-extrabold tracking-[-0.04em] text-[#bc0100]">
                YTDownloader
              </Link>
              <p className="text-xs uppercase tracking-[0.28em] text-[#956d67]">YouTube Downloader</p>
            </div>
          </div>

          <nav className="hidden items-center gap-2 md:flex">
            <Link
              to="/"
              className={`rounded-full px-4 py-2 text-sm font-semibold transition ${onHome ? "bg-[#bc0100] text-white" : "text-[#603e39] hover:bg-white"}`}
            >
              Search
            </Link>
            <a href="/#history" className="rounded-full px-4 py-2 text-sm font-semibold text-[#603e39] transition hover:bg-white">
              History
            </a>
            <Link
              to="/support"
              className={`rounded-full px-4 py-2 text-sm font-semibold transition ${onSupport ? "bg-[#bc0100] text-white" : "text-[#603e39] hover:bg-white"}`}
            >
              Support
            </Link>
          </nav>

          <div className="flex items-center gap-2">
            {authEnabled ? (
              <>
                <Show when="signed-out">
                  <SignInButton mode="modal">
                    <button
                      type="button"
                      onClick={() => trackEvent("sign_in_click", { source: "header" })}
                      className="rounded-full border border-[#ebbbb4] bg-white px-4 py-2 text-sm font-semibold text-[#603e39] transition hover:border-[#bc0100] hover:text-[#bc0100]"
                    >
                      Sign in
                    </button>
                  </SignInButton>
                  <SignUpButton mode="modal">
                    <button
                      type="button"
                      onClick={() => trackEvent("sign_up_click", { source: "header" })}
                      className="rounded-full bg-[#bc0100] px-4 py-2 text-sm font-semibold text-white shadow-[0_12px_28px_rgba(188,1,0,0.18)] transition hover:scale-[0.98]"
                    >
                      Create account
                    </button>
                  </SignUpButton>
                </Show>
                <Show when="signed-in">
                  <div className="flex items-center gap-3 rounded-full border border-white/70 bg-white/90 px-3 py-2 shadow-sm">
                    <span className="hidden text-sm font-medium text-[#603e39] sm:inline">Signed in</span>
                    <UserButton />
                  </div>
                </Show>
              </>
            ) : (
              <div className="rounded-full border border-[#ffdad6] bg-[#fff2f2] px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-[#930100]">
                Clerk not configured
              </div>
            )}
          </div>
        </div>
      </header>

      {children}

      <nav className="fixed inset-x-4 bottom-4 z-40 flex items-center justify-around rounded-[1.75rem] border border-white/70 bg-white/90 px-3 py-3 shadow-[0_14px_34px_rgba(28,27,27,0.08)] backdrop-blur-xl md:hidden">
        <Link className={`flex min-w-[72px] flex-col items-center gap-1 rounded-2xl px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] ${onHome ? "bg-[#fff1ef] text-[#bc0100]" : "text-[#956d67]"}`} to="/">
          <span className="material-symbols-outlined text-[20px]">search</span>
          Search
        </Link>
        <a className="flex min-w-[72px] flex-col items-center gap-1 rounded-2xl px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] text-[#956d67]" href="/#history">
          <span className="material-symbols-outlined text-[20px]">history</span>
          History
        </a>
        <Link
          className={`flex min-w-[72px] flex-col items-center gap-1 rounded-2xl px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] ${onSupport ? "bg-[#fff1ef] text-[#bc0100]" : "text-[#956d67]"}`}
          to="/support"
        >
          <span className="material-symbols-outlined text-[20px]">help</span>
          Support
        </Link>
        <div className={`flex min-w-[72px] flex-col items-center gap-1 rounded-2xl px-3 py-2 text-xs font-semibold uppercase tracking-[0.16em] ${onHistory ? "bg-[#fff1ef] text-[#bc0100]" : "text-[#956d67]"}`}>
          <span className="material-symbols-outlined text-[20px]">download</span>
          Job
        </div>
      </nav>
    </div>
  );
}

export default function App({ authEnabled = true }: AppProps) {
  return (
    <BrowserRouter>
      <AppShell authEnabled={authEnabled}>
        <RouteAnalytics />
        <Routes>
          <Route path="/" element={<HomePage authEnabled={authEnabled} />} />
          <Route path="/support" element={<SupportPage />} />
          <Route path="/downloads/:jobId" element={<DownloadPage authEnabled={authEnabled} />} />
          <Route path="*" element={<HomePage authEnabled={authEnabled} />} />
        </Routes>
      </AppShell>
    </BrowserRouter>
  );
}
