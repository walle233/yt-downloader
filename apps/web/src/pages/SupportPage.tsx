import { useState } from "react";
import { Link } from "react-router-dom";

const PURPOSE_POINTS = [
  {
    icon: "shield_lock",
    title: "Account-bound privacy",
    body: "Downloads, result links, and history stay tied to the signed-in account that created them.",
  },
  {
    icon: "video_settings",
    title: "Clear presets",
    body: "YTDownloader focuses on fixed presets instead of raw format lists, so you can choose quickly and predict the output.",
  },
] as const;

const FAQ_ITEMS = [
  {
    question: "Is YTDownloader free to use?",
    answer:
      "Yes. The current product flow includes anonymous probe and three free downloads per signed-in account. Subscription support is planned next, but it is not live yet.",
  },
  {
    question: "What links are supported right now?",
    answer:
      "This version is focused on YouTube only. Use a valid youtube.com/watch or youtu.be link when creating a download.",
  },
  {
    question: "What download formats are available?",
    answer:
      "YTDownloader currently offers MP4 presets in 360p, 480p, 720p, and 1080p, plus an audio-only MP3 option when the source supports it.",
  },
  {
    question: "Do I need to sign in before downloading?",
    answer:
      "Yes. Probing a link is public, but creating a download, viewing job details, opening result links, and accessing history all require sign-in.",
  },
  {
    question: "What happens after the three free downloads?",
    answer:
      "Once the three free downloads are used, the app will stop creating new download jobs on that account. The product will show a paywall-style notice, but checkout is not live in this version yet.",
  },
  {
    question: "How is my data handled?",
    answer:
      "We keep the product flow account-based so your downloads and history stay private to you. The app does not expose one user's jobs to another user's account.",
  },
] as const;

export function SupportPage() {
  const [openIndex, setOpenIndex] = useState<number>(0);

  return (
    <main className="pb-28">
      <section className="mx-auto grid w-full max-w-7xl grid-cols-1 gap-12 px-4 py-10 sm:px-6 lg:grid-cols-12 lg:items-start lg:gap-16 lg:py-20">
        <div className="space-y-6 lg:col-span-7">
          <p className="text-sm font-semibold uppercase tracking-[0.28em] text-[#bc0100]">The Curator&apos;s Guide</p>
          <div className="space-y-5">
            <h1 className="font-display max-w-4xl text-5xl font-extrabold tracking-[-0.06em] text-[#1c1b1b] sm:text-6xl md:text-7xl">
              YTDownloader support,
              <br />
              product truth, and FAQ.
            </h1>
            <p className="max-w-2xl text-lg leading-8 text-[#603e39]">
              YTDownloader is built to make YouTube downloading feel intentional instead of cluttered. This page explains what the product supports today, how the
              sign-in flow works, and what to expect from the download process.
            </p>
          </div>
        </div>

        <div className="relative lg:col-span-5">
          <div className="relative overflow-hidden rounded-[2.2rem] border border-white/80 bg-white/88 p-6 shadow-[0_20px_50px_rgba(28,27,27,0.08)]">
            <div className="absolute -right-12 -top-10 size-36 rounded-full bg-[#ffdad4]/80 blur-3xl" />
            <div className="relative aspect-square overflow-hidden rounded-[1.8rem] bg-[linear-gradient(145deg,_#f7f3f2_0%,_#ece4e3_100%)]">
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_25%,_rgba(188,1,0,0.12),_transparent_28%),radial-gradient(circle_at_70%_70%,_rgba(188,1,0,0.18),_transparent_24%)]" />
              <div className="absolute inset-6 rounded-[1.6rem] border border-white/70 bg-white/55 backdrop-blur-sm" />
              <div className="absolute inset-x-10 top-12 h-10 rounded-full bg-[#bc0100]/10" />
              <div className="absolute inset-x-14 top-28 h-4 rounded-full bg-[#bc0100]/12" />
              <div className="absolute inset-x-14 top-[9.5rem] h-4 rounded-full bg-[#bc0100]/8" />
              <div className="absolute bottom-10 right-10 grid size-28 place-items-center rounded-[1.8rem] bg-gradient-to-br from-[#bc0100] to-[#eb0000] text-white shadow-[0_18px_40px_rgba(188,1,0,0.22)]">
                <span className="material-symbols-outlined text-[38px]">help</span>
              </div>
            </div>
          </div>

          <div className="absolute -bottom-6 -left-2 max-w-xs rounded-[1.8rem] border border-[#ebbbb4]/40 bg-white px-6 py-5 shadow-[0_18px_36px_rgba(28,27,27,0.08)] sm:-left-6">
            <p className="font-display text-lg font-bold tracking-[-0.03em] text-[#1c1b1b]">Current product scope</p>
            <p className="mt-2 text-sm leading-6 text-[#603e39]">YouTube links, sign-in protected downloads, private history, fixed MP4 presets, and audio-only MP3.</p>
            <p className="mt-2 text-sm leading-6 text-[#603e39]">Each signed-in account currently includes three free downloads before the upcoming subscription layer.</p>
          </div>
        </div>
      </section>

      <section className="bg-[#f3efee]/80 py-20">
        <div className="mx-auto grid w-full max-w-7xl grid-cols-1 gap-12 px-4 sm:px-6 lg:grid-cols-[1.1fr_0.9fr] lg:items-center">
          <div className="space-y-6">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[#956d67]">About YTDownloader</p>
              <h2 className="mt-2 font-display text-4xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">A cleaner product layer on top of a familiar task.</h2>
            </div>
            <div className="space-y-5 text-lg leading-8 text-[#603e39]">
              <p>YTDownloader removes the usual downloader clutter and narrows the experience to a few predictable choices that match common personal-use needs.</p>
              <p>
                Instead of exposing every raw media variant, the app probes the video, shows only the presets that are truly available, and keeps completed jobs
                in a private history tied to your account, with a simple free plan before billing launches.
              </p>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            {PURPOSE_POINTS.map((item, index) => (
              <article
                key={item.title}
                className={`rounded-[1.8rem] border border-white/80 bg-white/92 p-6 shadow-[0_18px_40px_rgba(28,27,27,0.06)] ${index === 1 ? "sm:translate-y-8" : ""}`}
              >
                <div className="mb-4 grid size-12 place-items-center rounded-2xl bg-[#fff1ef] text-[#bc0100]">
                  <span className="material-symbols-outlined text-[24px]">{item.icon}</span>
                </div>
                <h3 className="font-display text-xl font-bold tracking-[-0.03em] text-[#1c1b1b]">{item.title}</h3>
                <p className="mt-3 text-sm leading-7 text-[#603e39]">{item.body}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="mx-auto w-full max-w-4xl px-4 py-20 sm:px-6">
        <div className="mb-10 text-center">
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[#956d67]">FAQ</p>
          <h2 className="mt-2 font-display text-4xl font-extrabold tracking-[-0.05em] text-[#1c1b1b]">Frequently asked</h2>
        </div>

        <div className="space-y-4">
          {FAQ_ITEMS.map((item, index) => {
            const isOpen = openIndex === index;
            return (
              <article key={item.question} className="overflow-hidden rounded-[1.5rem] border border-white/80 bg-white/88 shadow-[0_16px_40px_rgba(28,27,27,0.06)]">
                <button
                  type="button"
                  onClick={() => setOpenIndex((current) => (current === index ? -1 : index))}
                  className="flex w-full items-center justify-between gap-4 px-6 py-5 text-left"
                >
                  <span className="font-display text-xl font-bold tracking-[-0.03em] text-[#1c1b1b]">{item.question}</span>
                  <span className={`material-symbols-outlined text-[#bc0100] transition-transform ${isOpen ? "rotate-180" : ""}`}>expand_more</span>
                </button>
                <div className={`grid transition-[grid-template-rows] duration-300 ${isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]"}`}>
                  <div className="overflow-hidden">
                    <p className="px-6 pb-6 text-[15px] leading-7 text-[#603e39]">{item.answer}</p>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      </section>

      <section className="mx-auto w-full max-w-5xl px-4 pb-20 sm:px-6">
        <div className="relative overflow-hidden rounded-[2.3rem] bg-gradient-to-br from-[#eb0000] to-[#bc0100] px-8 py-10 text-white shadow-[0_22px_50px_rgba(188,1,0,0.2)] sm:px-12 sm:py-12">
          <div className="absolute right-0 top-0 size-56 -translate-y-1/3 translate-x-1/4 rounded-full bg-white/10 blur-3xl" />
          <div className="relative flex flex-col gap-8 md:flex-row md:items-center md:justify-between">
            <div className="max-w-2xl">
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-white/70">Need help</p>
              <h2 className="mt-2 font-display text-3xl font-extrabold tracking-[-0.04em]">Still deciding what to do next?</h2>
              <p className="mt-4 text-lg leading-8 text-white/82">
                If a job fails, start by verifying the YouTube URL, signing in again, and choosing one of the presets shown as available after probe.
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link
                to="/"
                className="inline-flex items-center justify-center gap-2 rounded-full bg-white px-6 py-4 text-sm font-bold uppercase tracking-[0.18em] text-[#bc0100] transition hover:scale-[0.98]"
              >
                <span className="material-symbols-outlined text-[18px]">travel_explore</span>
                Back to search
              </Link>
              <a
                href="/#history"
                className="inline-flex items-center justify-center gap-2 rounded-full border border-white/25 bg-white/8 px-6 py-4 text-sm font-bold uppercase tracking-[0.18em] text-white transition hover:bg-white/12"
              >
                <span className="material-symbols-outlined text-[18px]">history</span>
                View history
              </a>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
