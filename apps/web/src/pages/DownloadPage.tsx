import type { ReactNode } from "react";
import { SignInButton, useAuth } from "@clerk/react";
import type { DownloadResultResponse, DownloadStatusResponse } from "@ytvd/shared-types";
import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { APIError, authorizedRequestJSON } from "../lib/api";
import { trackEvent } from "../lib/analytics";
import { formatBytes, formatDuration, profileHeadline, profileSubline, statusLabel } from "../lib/format";

export function DownloadPage({ authEnabled }: { authEnabled: boolean }) {
  const { jobId = "" } = useParams();
  const { getToken, isSignedIn } = useAuth();

  const [status, setStatus] = useState<DownloadStatusResponse | null>(null);
  const [result, setResult] = useState<DownloadResultResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const hasTrackedReadyRef = useRef(false);

  useEffect(() => {
    hasTrackedReadyRef.current = false;
  }, [jobId]);

  useEffect(() => {
    if (!jobId || !authEnabled || !isSignedIn) {
      return;
    }

    let cancelled = false;
    let intervalId: number | undefined;

    const load = async () => {
      setLoading(true);
      try {
        const statusData = await authorizedRequestJSON<DownloadStatusResponse>(`/downloads/${jobId}`, getToken);
        if (cancelled) {
          return;
        }
        setStatus(statusData);

        if (statusData.status === "completed") {
          const resultData = await authorizedRequestJSON<DownloadResultResponse>(`/downloads/${jobId}/result`, getToken);
          if (!cancelled) {
            setResult(resultData);
            if (!hasTrackedReadyRef.current) {
              trackEvent("download_result_ready", {
                profile_id: statusData.profileId,
                media_kind: statusData.mediaKind,
              });
              hasTrackedReadyRef.current = true;
            }
          }
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load job");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void load();
    intervalId = window.setInterval(() => {
      void load();
    }, 3000);

    return () => {
      cancelled = true;
      if (intervalId) {
        window.clearInterval(intervalId);
      }
    };
  }, [authEnabled, getToken, isSignedIn, jobId]);

  if (!authEnabled) {
    return <GateCard title="Clerk is not configured." body="Add the publishable key runtime config to the frontend and a backend auth key before using protected download pages." />;
  }

  if (!isSignedIn) {
    return (
      <GateCard
        title="Sign in to view this download."
        body="Download details, result links, and your private history are available only after authentication."
        action={
          <SignInButton mode="modal">
            <button type="button" className="rounded-full bg-[#bc0100] px-5 py-3 text-sm font-semibold text-white">
              Sign in
            </button>
          </SignInButton>
        }
      />
    );
  }

  if (error) {
    return <GateCard title="This job is not available." body={error} />;
  }

  if (!status) {
    return <GateCard title="Loading download details..." body={loading ? "Fetching the current job state from the API." : "Preparing the page."} />;
  }

  return (
    <main className="mx-auto w-full max-w-6xl px-4 pb-28 pt-10 sm:px-6">
      <section className="grid gap-10 lg:grid-cols-[1.05fr_0.95fr] lg:items-start">
        <div className="space-y-6">
          <Link to="/" className="inline-flex items-center gap-2 rounded-full border border-[#ebbbb4] bg-white px-4 py-2 text-sm font-semibold text-[#603e39] transition hover:border-[#bc0100] hover:text-[#bc0100]">
            <span className="material-symbols-outlined text-[18px]">arrow_back</span>
            Back to search
          </Link>

          <div className="overflow-hidden rounded-[2.4rem] border border-white/80 bg-white/88 shadow-[0_20px_50px_rgba(28,27,27,0.08)]">
            <div className="relative aspect-video bg-[#f0edec]">
              {status.thumbnailUrl ? <img src={status.thumbnailUrl} alt={status.title} className="h-full w-full object-cover" /> : null}
              <div className="absolute inset-0 bg-gradient-to-t from-black/75 to-transparent" />
              <div className="absolute inset-x-0 bottom-0 p-8 text-white">
                <span className="rounded-full bg-white/20 px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em]">{statusLabel(status.status)}</span>
                <h1 className="mt-4 font-display text-3xl font-extrabold tracking-[-0.05em] sm:text-4xl">{status.title}</h1>
                <p className="mt-2 text-sm text-white/80">Duration {formatDuration(status.durationSec)}</p>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-[2rem] border border-white/80 bg-white/88 p-6 shadow-[0_18px_44px_rgba(28,27,27,0.07)] sm:p-8">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.22em] text-[#956d67]">Current job</p>
                <h2 className="mt-2 font-display text-3xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">{profileHeadline(status.profileId, status.targetHeight)}</h2>
              </div>
              <span className="rounded-full bg-[#fff1ef] px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] text-[#bc0100]">{statusLabel(status.status)}</span>
            </div>

            <div className="mt-6 space-y-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <Metric label="Preset" value={profileSubline(status.mediaKind, status.targetHeight)} />
                <Metric label="Created" value={new Date(status.createdAt).toLocaleString()} />
                <Metric label="Progress" value={`${status.progress}%`} />
                <Metric label="Job ID" value={status.jobId} />
              </div>

              <div>
                <div className="h-3 overflow-hidden rounded-full bg-[#f0edec]">
                  <div className="h-full rounded-full bg-gradient-to-r from-[#bc0100] to-[#eb0000]" style={{ width: `${Math.max(status.progress, 6)}%` }} />
                </div>
                <p className="mt-3 text-sm text-[#603e39]">{status.message || `${statusLabel(status.status)} • ${status.step}`}</p>
              </div>
            </div>
          </div>

          <div className="rounded-[2rem] border border-white/80 bg-white/88 p-6 shadow-[0_18px_44px_rgba(28,27,27,0.07)] sm:p-8">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-[#956d67]">Delivery</p>
            {result ? (
              <div className="mt-3 space-y-4">
                <div>
                  <h3 className="font-display text-2xl font-extrabold tracking-[-0.04em] text-[#1c1b1b]">Your file is ready.</h3>
                  <p className="mt-2 text-[#603e39]">
                    {result.fileName} • {formatBytes(result.fileSize)}
                  </p>
                </div>
                <a
                  onClick={() =>
                    trackEvent("download_link_open", {
                      profile_id: status.profileId,
                      media_kind: status.mediaKind,
                    })
                  }
                  className="inline-flex w-full items-center justify-center gap-2 rounded-[1.2rem] bg-[#bc0100] px-6 py-4 text-sm font-bold uppercase tracking-[0.18em] text-white shadow-[0_18px_38px_rgba(188,1,0,0.22)] transition hover:scale-[0.99]"
                  href={result.downloadUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  <span className="material-symbols-outlined text-[20px]">download</span>
                  Open download link
                </a>
              </div>
            ) : (
              <div className="mt-3 space-y-3 text-[#603e39]">
                <h3 className="font-display text-2xl font-extrabold tracking-[-0.04em] text-[#1c1b1b]">We are still processing this file.</h3>
                <p>The result link will appear here automatically once packaging is complete.</p>
              </div>
            )}
          </div>
        </div>
      </section>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-[1.4rem] bg-[#fcf9f8] px-4 py-4">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-[#956d67]">{label}</p>
      <p className="mt-2 break-all text-sm font-medium text-[#1c1b1b]">{value}</p>
    </div>
  );
}

function GateCard({ title, body, action }: { title: string; body: string; action?: ReactNode }) {
  return (
    <main className="mx-auto flex min-h-[70vh] w-full max-w-5xl items-center px-4 py-12 sm:px-6">
      <div className="w-full rounded-[2rem] border border-white/80 bg-white/88 p-8 shadow-[0_18px_44px_rgba(28,27,27,0.07)] sm:p-10">
        <p className="text-xs font-semibold uppercase tracking-[0.22em] text-[#956d67]">Archive</p>
        <h1 className="mt-3 font-display text-4xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">{title}</h1>
        <p className="mt-4 max-w-2xl text-lg leading-8 text-[#603e39]">{body}</p>
        {action ? <div className="mt-8">{action}</div> : null}
      </div>
    </main>
  );
}
