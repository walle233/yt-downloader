type AnalyticsParams = Record<string, string | number | boolean | null | undefined>;
const HARDCODED_MEASUREMENT_ID = "G-1TD283G0DN";

declare global {
  interface Window {
    dataLayer?: unknown[];
    gtag?: (...args: unknown[]) => void;
    google_tag_manager?: Record<string, unknown>;
  }
}

let initialized = false;

export function getAnalyticsMeasurementId(): string {
  return window.__RUNTIME_CONFIG__?.VITE_GA_MEASUREMENT_ID || import.meta.env.VITE_GA_MEASUREMENT_ID || HARDCODED_MEASUREMENT_ID;
}

export function initAnalytics(): boolean {
  const measurementId = getAnalyticsMeasurementId();
  if (!measurementId) {
    return false;
  }

  const existingScript = document.querySelector<HTMLScriptElement>(
    `script[src="https://www.googletagmanager.com/gtag/js?id=${measurementId}"]`,
  );
  const existingTagManager = Boolean(window.google_tag_manager?.[measurementId]);

  if ((existingScript || existingTagManager) && typeof window.gtag === "function") {
    initialized = true;
    return true;
  }

  if (!window.dataLayer) {
    window.dataLayer = [];
  }

  if (!window.gtag) {
    window.gtag = function gtag(...args: unknown[]) {
      window.dataLayer?.push(args);
    };
  }

  if (!document.querySelector(`script[data-ga-id="${measurementId}"]`)) {
    const script = document.createElement("script");
    script.async = true;
    script.src = `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(measurementId)}`;
    script.dataset.gaId = measurementId;
    document.head.appendChild(script);
  }

  if (!initialized) {
    window.gtag("js", new Date());
    window.gtag("config", measurementId, { send_page_view: false });
    initialized = true;
  }

  return true;
}

export function trackPageView(path: string, title?: string) {
  const measurementId = getAnalyticsMeasurementId();
  if (!measurementId || !initAnalytics()) {
    return;
  }

  window.gtag?.("event", "page_view", {
    page_title: title || document.title,
    page_path: path,
    page_location: window.location.href,
  });
}

export function trackEvent(name: string, params?: AnalyticsParams) {
  if (!initAnalytics()) {
    return;
  }

  window.gtag?.("event", name, params || {});
}
