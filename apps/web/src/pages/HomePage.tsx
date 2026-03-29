import { Show, SignInButton, SignUpButton, useAuth, useClerk } from "@clerk/react";
import type {
  BillingResponse,
  BillingSummary,
  CreateDownloadResponse,
  DownloadListItem,
  DownloadListResponse,
  DownloadProfile,
  ProbeResponse,
} from "@ytvd/shared-types";
import { startTransition, useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { APIError, authorizedRequestJSON, isFreeLimitErrorResponse, requestJSON } from "../lib/api";
import { trackEvent } from "../lib/analytics";
import { formatBytes, formatDuration, groupProfiles, profileHeadline, profileSubline, sortProfiles, statusLabel } from "../lib/format";

const DEFAULT_BILLING: BillingSummary = {
  plan: "free",
  subscriptionStatus: "inactive",
  freeDownloadsLimit: 3,
  freeDownloadsUsed: 0,
  freeDownloadsRemaining: 3,
  canDownload: true,
};

export function HomePage({ authEnabled }: { authEnabled: boolean }) {
  const navigate = useNavigate();
  const { getToken, isSignedIn } = useAuth();
  const { openSignIn } = useClerk();

  const [url, setURL] = useState("https://www.youtube.com/watch?v=2zda1Tr4big");
  const [probe, setProbe] = useState<ProbeResponse | null>(null);
  const [history, setHistory] = useState<DownloadListItem[]>([]);
  const [billing, setBilling] = useState<BillingSummary>(DEFAULT_BILLING);
  const [loadingProbe, setLoadingProbe] = useState(false);
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [creatingProfileId, setCreatingProfileId] = useState<string | null>(null);
  const [showPaywall, setShowPaywall] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    if (!isSignedIn) {
      setHistory([]);
      setBilling(DEFAULT_BILLING);
      setShowPaywall(false);
      return;
    }
    void loadSignedInState();
  }, [authEnabled, isSignedIn]);

  const loadSignedInState = async () => {
    setLoadingHistory(true);

    try {
      const [downloadsData, billingData] = await Promise.all([
        authorizedRequestJSON<DownloadListResponse>("/downloads", getToken),
        authorizedRequestJSON<BillingResponse>("/billing", getToken),
      ]);
      startTransition(() => {
        setHistory(downloadsData.items);
        setBilling(billingData.billing);
        setShowPaywall(!billingData.billing.canDownload);
      });
    } catch (err) {
      if (err instanceof APIError && err.status === 401) {
        startTransition(() => {
          setHistory([]);
          setBilling(DEFAULT_BILLING);
          setShowPaywall(false);
        });
        return;
      }
      setError(err instanceof Error ? err.message : "Failed to load history");
    } finally {
      setLoadingHistory(false);
    }
  };

  const handleProbe = async () => {
    setLoadingProbe(true);
    setError(null);
    setNotice(null);
    trackEvent("probe_submit", {
      signed_in: !!isSignedIn,
      url_host: safeURLHost(url),
    });

    try {
      const data = await requestJSON<ProbeResponse>("/videos/probe", {
        method: "POST",
        body: JSON.stringify({ url }),
      });
      startTransition(() => setProbe({ ...data, profiles: sortProfiles(data.profiles) }));
      trackEvent("probe_success", {
        video_profiles: data.profiles.filter((profile) => profile.kind === "video" && profile.available).length,
        audio_profiles: data.profiles.filter((profile) => profile.kind === "audio" && profile.available).length,
      });
      setNotice("Review the available presets below. Sign in when you are ready to use one of your three free downloads.");
    } catch (err) {
      trackEvent("probe_failed", {
        signed_in: !!isSignedIn,
        message: err instanceof Error ? err.message.slice(0, 120) : "unknown_error",
      });
      setError(err instanceof Error ? err.message : "Failed to analyze the video");
    } finally {
      setLoadingProbe(false);
    }
  };

  const handleCreate = async (profile: DownloadProfile) => {
    if (!profile.available) {
      return;
    }
    trackEvent("download_preset_click", {
      profile_id: profile.id,
      media_kind: profile.kind,
      signed_in: !!isSignedIn,
    });
    if (!authEnabled) {
      setError("Clerk is not configured in this environment.");
      return;
    }
    if (!isSignedIn) {
      trackEvent("sign_in_click", { source: "preset_gate", profile_id: profile.id });
      await openSignIn({
        afterSignInUrl: window.location.href,
        afterSignUpUrl: window.location.href,
      });
      return;
    }
    if (!billing.canDownload) {
      setShowPaywall(true);
      setNotice("Your three free downloads have been used. Subscription support is the next upgrade on the roadmap.");
      return;
    }

    setCreatingProfileId(profile.id);
    setError(null);
    setNotice(null);

    try {
      const data = await authorizedRequestJSON<CreateDownloadResponse>("/downloads", getToken, {
        method: "POST",
        body: JSON.stringify({ url, profileId: profile.id }),
      });
      trackEvent("download_created", {
        profile_id: profile.id,
        media_kind: profile.kind,
      });
      navigate(`/downloads/${data.jobId}`);
    } catch (err) {
      const errorData = err instanceof APIError ? err.data : undefined;
      if (isFreeLimitErrorResponse(errorData)) {
        startTransition(() => {
          setBilling(errorData.billing);
          setShowPaywall(true);
        });
        setNotice("You have used all three free downloads on this account. Billing support is the next upgrade and is not live yet.");
        return;
      }
      trackEvent("download_create_failed", {
        profile_id: profile.id,
        media_kind: profile.kind,
        message: err instanceof Error ? err.message.slice(0, 120) : "unknown_error",
      });
      setError(err instanceof Error ? err.message : "Failed to create download");
    } finally {
      setCreatingProfileId(null);
    }
  };

  const groupedProfiles = probe ? groupProfiles(probe.profiles) : { video: [], audio: [] };

  return (
    <main className="pb-28">
      <section className="mx-auto grid w-full max-w-7xl grid-cols-1 gap-10 px-4 py-10 sm:px-6 lg:grid-cols-12 lg:items-center lg:gap-16 lg:py-20">
        <div className="space-y-6 lg:col-span-7">
          <span className="inline-flex rounded-full border border-[#ebbbb4] bg-white/80 px-4 py-2 text-xs font-semibold uppercase tracking-[0.28em] text-[#930100]">
            Product-ready YouTube downloads
          </span>
          <div className="space-y-4">
            <h1 className="font-display max-w-3xl text-5xl font-extrabold tracking-[-0.06em] text-[#1c1b1b] sm:text-6xl md:text-7xl">
              Download YouTube video or audio with a clean, reliable flow.
            </h1>
            <p className="max-w-2xl text-lg leading-8 text-[#603e39]">
              Analyze a link instantly, use up to three free downloads per account, and keep every completed file in your private library.
            </p>
          </div>

          <div className="flex flex-wrap gap-4 text-sm font-medium text-[#603e39]">
            <div className="inline-flex items-center gap-2 rounded-full bg-white/80 px-4 py-2 shadow-sm">
              <span className="material-symbols-outlined text-[18px] text-[#bc0100]">check_circle</span>
              360p / 480p / 720p / 1080p
            </div>
            <div className="inline-flex items-center gap-2 rounded-full bg-white/80 px-4 py-2 shadow-sm">
              <span className="material-symbols-outlined text-[18px] text-[#bc0100]">headphones</span>
              Audio-only MP3
            </div>
            <div className="inline-flex items-center gap-2 rounded-full bg-white/80 px-4 py-2 shadow-sm">
              <span className="material-symbols-outlined text-[18px] text-[#bc0100]">shield_lock</span>
              3 free downloads per account
            </div>
          </div>

          <Show when="signed-out">
            <div className="flex flex-wrap gap-3">
              <SignInButton mode="modal">
                <button
                  type="button"
                  onClick={() => trackEvent("sign_in_click", { source: "hero" })}
                  className="inline-flex items-center gap-2 rounded-full bg-[#bc0100] px-5 py-3 text-sm font-semibold text-white shadow-[0_16px_34px_rgba(188,1,0,0.2)] transition hover:scale-[0.98]"
                >
                  <span className="material-symbols-outlined text-[18px]">login</span>
                  Sign in to unlock downloads
                </button>
              </SignInButton>
              <SignUpButton mode="modal">
                <button
                  type="button"
                  onClick={() => trackEvent("sign_up_click", { source: "hero" })}
                  className="inline-flex items-center gap-2 rounded-full border border-[#ebbbb4] bg-white px-5 py-3 text-sm font-semibold text-[#603e39] transition hover:border-[#bc0100] hover:text-[#bc0100]"
                >
                  <span className="material-symbols-outlined text-[18px]">person_add</span>
                  Create free account
                </button>
              </SignUpButton>
            </div>
          </Show>
        </div>

        <div className="relative lg:col-span-5">
          <div className="absolute -bottom-4 -right-4 hidden h-full w-full rounded-[2rem] bg-[#f0edec] lg:block" />
          <div className="relative overflow-hidden rounded-[2rem] border border-white/80 bg-white/88 p-6 shadow-[0_20px_50px_rgba(28,27,27,0.08)] backdrop-blur-xl sm:p-8">
            <div className="absolute -right-16 -top-16 size-40 rounded-full bg-[#ffdad4]/70 blur-3xl" />
            <div className="relative space-y-6">
              <div className="space-y-2">
                <p className="text-sm font-semibold uppercase tracking-[0.22em] text-[#956d67]">Paste YouTube URL</p>
                <label className="block overflow-hidden rounded-[1.4rem] border border-[#ebbbb4]/60 bg-[#fcf9f8] shadow-sm transition focus-within:border-[#bc0100] focus-within:bg-white">
                  <textarea
                    className="min-h-[132px] w-full resize-none border-0 bg-transparent px-5 py-5 text-base leading-7 text-[#1c1b1b] outline-none placeholder:text-[#9d8b88]"
                    value={url}
                    onChange={(event) => setURL(event.target.value)}
                    placeholder="https://youtube.com/watch?v=..."
                  />
                </label>
              </div>

              <button
                type="button"
                onClick={handleProbe}
                disabled={loadingProbe}
                className="inline-flex w-full items-center justify-center gap-3 rounded-[1.2rem] bg-gradient-to-br from-[#bc0100] to-[#eb0000] px-6 py-4 text-sm font-bold uppercase tracking-[0.18em] text-white shadow-[0_18px_38px_rgba(188,1,0,0.22)] transition hover:scale-[0.99] disabled:cursor-default disabled:opacity-60"
              >
                <span className="material-symbols-outlined text-[20px]">travel_explore</span>
                {loadingProbe ? "Analyzing..." : "Analyze video"}
              </button>

              <div className="grid gap-3 text-sm text-[#603e39] sm:grid-cols-3">
                <div className="rounded-2xl bg-[#fcf9f8] px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.2em] text-[#956d67]">Access</p>
                  <p className="mt-1 font-display text-2xl font-extrabold tracking-[-0.04em]">{isSignedIn && billing.plan === "pro" && billing.canDownload ? "Pro" : isSignedIn ? billing.freeDownloadsRemaining : 3}</p>
                  <p className="text-xs">
                    {isSignedIn && billing.plan === "pro" && billing.canDownload ? "Active subscription" : isSignedIn ? "Free downloads left" : "Downloads after sign-in"}
                  </p>
                </div>
                <div className="rounded-2xl bg-[#fcf9f8] px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.2em] text-[#956d67]">Formats</p>
                  <p className="mt-1 font-display text-2xl font-extrabold tracking-[-0.04em]">5</p>
                  <p className="text-xs">Curated presets</p>
                </div>
                <div className="rounded-2xl bg-[#fcf9f8] px-4 py-3">
                  <p className="text-xs uppercase tracking-[0.2em] text-[#956d67]">Library</p>
                  <p className="mt-1 font-display text-2xl font-extrabold tracking-[-0.04em]">{history.length}</p>
                  <p className="text-xs">Recent jobs</p>
                </div>
              </div>

              {notice ? <p className="rounded-2xl bg-[#fff4f2] px-4 py-3 text-sm text-[#6a3b34]">{notice}</p> : null}
              {error ? <p className="rounded-2xl bg-[#fff0ee] px-4 py-3 text-sm text-[#930100]">{error}</p> : null}
            </div>
          </div>
        </div>
      </section>

      {probe ? (
        <section className="mx-auto w-full max-w-7xl px-4 pb-6 sm:px-6">
          <div className="grid gap-8 lg:grid-cols-[1.15fr_0.85fr]">
            <div className="overflow-hidden rounded-[2.25rem] border border-white/80 bg-white/85 shadow-[0_20px_50px_rgba(28,27,27,0.08)]">
              <div className="grid gap-0 lg:grid-cols-[1.1fr_0.9fr]">
                <div className="relative aspect-video bg-[#f0edec]">
                  {probe.thumbnailUrl ? <img src={probe.thumbnailUrl} alt={probe.title} className="h-full w-full object-cover" /> : null}
                  <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/75 to-transparent p-6 text-white">
                    <span className="rounded-full bg-white/20 px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em]">YouTube</span>
                    <h2 className="mt-3 font-display text-2xl font-extrabold tracking-[-0.04em] sm:text-3xl">{probe.title}</h2>
                    <p className="mt-2 text-sm text-white/80">Duration {formatDuration(probe.durationSec)}</p>
                  </div>
                </div>
                <div className="space-y-4 p-6 sm:p-8">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#956d67]">Download presets</p>
                    <h3 className="mt-2 font-display text-3xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">Choose your output</h3>
                  </div>
                  <p className="text-sm leading-7 text-[#603e39]">
                    Probe is public. Creating a real download is protected behind your account and counts toward the three free downloads available on the free plan.
                  </p>
                  <div className="rounded-[1.5rem] bg-[#fcf9f8] p-4 text-sm text-[#603e39]">
                    <p className="font-semibold text-[#1c1b1b]">What is included</p>
                    <ul className="mt-3 space-y-2">
                      <li className="flex items-start gap-2">
                        <span className="material-symbols-outlined mt-0.5 text-[18px] text-[#bc0100]">check_circle</span>
                        Exact product presets instead of raw source dumps
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="material-symbols-outlined mt-0.5 text-[18px] text-[#bc0100]">check_circle</span>
                        Private history for the signed-in account only
                      </li>
                      <li className="flex items-start gap-2">
                        <span className="material-symbols-outlined mt-0.5 text-[18px] text-[#bc0100]">check_circle</span>
                        Audio-only MP3 option for music, podcasts, and speeches
                      </li>
                    </ul>
                  </div>
                  {showPaywall ? (
                    <PaywallCard
                      billing={billing}
                      onDismiss={() => {
                        setShowPaywall(false);
                        setNotice("Subscription support is not live yet. You can still review your history and come back after the billing upgrade ships.");
                      }}
                    />
                  ) : null}
                </div>
              </div>
            </div>

            <div className="space-y-6">
              <ProfileGroup
                title="Video presets"
                subtitle="MP4 output"
                profiles={groupedProfiles.video}
                creatingProfileId={creatingProfileId}
                isSignedIn={!!isSignedIn}
                canDownload={billing.canDownload}
                onCreate={handleCreate}
              />
              <ProfileGroup
                title="Audio-only"
                subtitle="MP3 output"
                profiles={groupedProfiles.audio}
                creatingProfileId={creatingProfileId}
                isSignedIn={!!isSignedIn}
                canDownload={billing.canDownload}
                onCreate={handleCreate}
              />
            </div>
          </div>
        </section>
      ) : null}

      <section id="history" className="mx-auto w-full max-w-7xl px-4 py-14 sm:px-6">
        <div className="mb-6 flex items-end justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[#956d67]">Private library</p>
            <h2 className="mt-2 font-display text-3xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">Recent downloads</h2>
          </div>
              {isSignedIn ? (
            <button
              type="button"
              className="rounded-full border border-[#ebbbb4] bg-white px-4 py-2 text-sm font-semibold text-[#603e39] transition hover:border-[#bc0100] hover:text-[#bc0100]"
              onClick={() => void loadSignedInState()}
            >
              {loadingHistory ? "Refreshing..." : "Refresh"}
            </button>
          ) : null}
        </div>

        <Show when="signed-out">
          <div className="rounded-[2rem] border border-white/80 bg-white/88 p-8 shadow-[0_18px_44px_rgba(28,27,27,0.07)]">
            <div className="max-w-2xl space-y-4">
              <h3 className="font-display text-3xl font-extrabold tracking-[-0.04em] text-[#1c1b1b]">Sign in to create downloads and keep history.</h3>
              <p className="text-[#603e39]">
                Probe any YouTube link anonymously, then sign in to use your three free downloads and keep a private account-bound history.
              </p>
              <div className="flex flex-wrap gap-3">
                <SignInButton mode="modal">
                  <button type="button" onClick={() => trackEvent("sign_in_click", { source: "history_gate" })} className="rounded-full bg-[#bc0100] px-5 py-3 text-sm font-semibold text-white">
                    Sign in
                  </button>
                </SignInButton>
                <SignUpButton mode="modal">
                  <button type="button" onClick={() => trackEvent("sign_up_click", { source: "history_gate" })} className="rounded-full border border-[#ebbbb4] bg-white px-5 py-3 text-sm font-semibold text-[#603e39]">
                    Create free account
                  </button>
                </SignUpButton>
              </div>
            </div>
          </div>
        </Show>

        <Show when="signed-in">
          {history.length > 0 ? (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {history.map((item) => (
                <Link
                  key={item.jobId}
                  to={`/downloads/${item.jobId}`}
                  className="group overflow-hidden rounded-[1.8rem] border border-white/80 bg-white/88 shadow-[0_18px_44px_rgba(28,27,27,0.07)] transition hover:-translate-y-1"
                >
                  <div className="relative aspect-video bg-[#f0edec]">
                    {item.thumbnailUrl ? <img src={item.thumbnailUrl} alt={item.title} className="h-full w-full object-cover transition duration-500 group-hover:scale-105" /> : null}
                    <div className="absolute inset-0 bg-gradient-to-t from-black/70 to-transparent" />
                    <div className="absolute inset-x-0 bottom-0 p-5 text-white">
                      <div className="inline-flex rounded-full bg-white/20 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em]">
                        {profileHeadline(item.profileId, item.targetHeight)}
                      </div>
                      <p className="mt-3 line-clamp-2 font-display text-xl font-bold tracking-[-0.03em]">{item.title}</p>
                    </div>
                  </div>
                  <div className="flex items-center justify-between px-5 py-4 text-sm text-[#603e39]">
                    <span>{statusLabel(item.status)}</span>
                    <span>{new Date(item.createdAt).toLocaleString()}</span>
                  </div>
                </Link>
              ))}
            </div>
          ) : (
            <div className="rounded-[2rem] border border-white/80 bg-white/88 p-8 shadow-[0_18px_44px_rgba(28,27,27,0.07)]">
              <h3 className="font-display text-2xl font-bold tracking-[-0.04em] text-[#1c1b1b]">Your library is empty.</h3>
              <p className="mt-3 max-w-2xl text-[#603e39]">
                Analyze a YouTube link above, pick an available preset, and your completed downloads will start appearing here automatically.
              </p>
            </div>
          )}
        </Show>
      </section>

      <section id="support" className="mx-auto w-full max-w-7xl px-4 pb-20 sm:px-6">
        <div className="rounded-[2rem] border border-white/80 bg-white/88 p-8 shadow-[0_18px_44px_rgba(28,27,27,0.07)]">
          <div className="grid gap-6 md:grid-cols-[1fr_auto] md:items-center">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.22em] text-[#956d67]">Support</p>
              <h2 className="mt-2 font-display text-3xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">Built for clean, private personal use.</h2>
              <p className="mt-3 max-w-3xl text-[#603e39]">
                YTDownloader is focused on YouTube downloads with clear presets, account-bound history, three free downloads per account, and subscription support coming next.
              </p>
            </div>
            <div className="rounded-[1.5rem] bg-[#fcf9f8] px-5 py-4 text-sm text-[#603e39]">
              <p className="font-semibold text-[#1c1b1b]">Need help?</p>
              <p className="mt-1">Use a valid `youtube.com/watch` or `youtu.be` link and sign in before creating the download.</p>
              <Link to="/support" className="mt-4 inline-flex items-center gap-2 font-semibold text-[#bc0100] transition hover:text-[#930100]">
                Open support guide
                <span className="material-symbols-outlined text-[18px]">arrow_forward</span>
              </Link>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}

function safeURLHost(value: string) {
  try {
    return new URL(value).host;
  } catch {
    return "invalid_url";
  }
}

function PaywallCard({ billing, onDismiss }: { billing: BillingSummary; onDismiss: () => void }) {
  return (
    <div className="rounded-[1.6rem] border border-[#ffdad6] bg-[#fff3f1] p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#930100]">Free plan limit reached</p>
          <h4 className="mt-2 font-display text-2xl font-extrabold tracking-[-0.04em] text-[#1c1b1b]">Your free downloads are used up.</h4>
        </div>
        <button type="button" onClick={onDismiss} className="rounded-full border border-[#ebbbb4] bg-white px-3 py-1 text-xs font-semibold uppercase tracking-[0.16em] text-[#603e39]">
          Dismiss
        </button>
      </div>
      <p className="mt-3 text-sm leading-7 text-[#6a3b34]">
        You have used {billing.freeDownloadsUsed} of {billing.freeDownloadsLimit} free downloads on this account. Subscription support is the next feature on the roadmap, so there is no checkout flow yet.
      </p>
      <div className="mt-4 flex flex-wrap gap-3">
        <Link to="/support" className="inline-flex items-center gap-2 rounded-full bg-[#bc0100] px-4 py-3 text-sm font-semibold text-white">
          <span className="material-symbols-outlined text-[18px]">help</span>
          Read the support guide
        </Link>
        <span className="inline-flex items-center gap-2 rounded-full border border-[#ebbbb4] bg-white px-4 py-3 text-sm font-semibold text-[#603e39]">
          <span className="material-symbols-outlined text-[18px]">hourglass_top</span>
          Subscription coming soon
        </span>
      </div>
    </div>
  );
}

function ProfileGroup({
  title,
  subtitle,
  profiles,
  creatingProfileId,
  isSignedIn,
  canDownload,
  onCreate,
}: {
  title: string;
  subtitle: string;
  profiles: DownloadProfile[];
  creatingProfileId: string | null;
  isSignedIn: boolean;
  canDownload: boolean;
  onCreate: (profile: DownloadProfile) => Promise<void>;
}) {
  if (profiles.length === 0) {
    return null;
  }

  return (
    <section className="rounded-[2rem] border border-white/80 bg-white/88 p-6 shadow-[0_18px_44px_rgba(28,27,27,0.07)] sm:p-7">
      <div className="mb-5 flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-[#956d67]">{subtitle}</p>
          <h3 className="mt-2 font-display text-2xl font-extrabold tracking-[-0.04em] text-[#1c1b1b]">{title}</h3>
        </div>
      </div>
      <div className="space-y-4">
        {profiles.map((profile) => {
          const isBusy = creatingProfileId === profile.id;
          return (
            <button
              key={profile.id}
              type="button"
              disabled={!profile.available || isBusy}
              onClick={() => void onCreate(profile)}
              className={`flex w-full items-center justify-between rounded-[1.6rem] border px-5 py-5 text-left transition ${
                profile.available
                  ? "border-[#ebbbb4]/50 bg-[#fcf9f8] hover:border-[#bc0100] hover:bg-white"
                  : "cursor-not-allowed border-[#ebe7e7] bg-[#f5f1f0] opacity-60"
              }`}
            >
              <div className="flex items-center gap-4">
                <div className={`grid size-12 place-items-center rounded-full ${profile.kind === "audio" ? "bg-[#fff0ee] text-[#bc0100]" : "bg-white text-[#bc0100]"}`}>
                  <span className="material-symbols-outlined text-[22px]">{profile.kind === "audio" ? "headphones" : "videocam"}</span>
                </div>
                <div>
                  <p className="font-display text-xl font-bold tracking-[-0.03em] text-[#1c1b1b]">{profileHeadline(profile.id, profile.targetHeight)}</p>
                  <p className="text-sm text-[#603e39]">
                    {profileSubline(profile.kind, profile.targetHeight)} • {formatBytes(profile.estimatedSizeBytes)}
                  </p>
                </div>
              </div>
              <div className="text-right">
                <span className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-[0.18em] ${profile.available ? "bg-[#fff1ef] text-[#bc0100]" : "bg-[#ebe7e7] text-[#7d7674]"}`}>
                  {profile.available ? (isSignedIn ? (canDownload ? (isBusy ? "Creating" : "Download") : "Limit reached") : "Sign in") : "Unavailable"}
                </span>
              </div>
            </button>
          );
        })}
      </div>
    </section>
  );
}
