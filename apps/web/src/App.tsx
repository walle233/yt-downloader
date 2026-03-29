import { useEffect, useRef, useState } from "react";
import { Show, SignInButton, SignUpButton, UserButton } from "@clerk/react";
import type {
  CreateDownloadResponse,
  DownloadListItem,
  DownloadListResponse,
  DownloadResultResponse,
  DownloadStatusResponse,
  OutputFormat,
  ProbeResponse,
} from "@ytvd/shared-types";

const API_BASE = "/api/v1";

async function requestJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    const fallback = await response.text();
    throw new Error(fallback || `Request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

interface AppProps {
  authEnabled?: boolean;
}

export default function App({ authEnabled = true }: AppProps) {
  const [url, setURL] = useState("https://www.youtube.com/watch?v=fiVdZ3ZkIjw");
  const [format, setFormat] = useState<OutputFormat>("mp4");
  const [probe, setProbe] = useState<ProbeResponse | null>(null);
  const [job, setJob] = useState<CreateDownloadResponse | null>(null);
  const [status, setStatus] = useState<DownloadStatusResponse | null>(null);
  const [result, setResult] = useState<DownloadResultResponse | null>(null);
  const [history, setHistory] = useState<DownloadListItem[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [loadingProbe, setLoadingProbe] = useState(false);
  const [loadingCreate, setLoadingCreate] = useState(false);
  const [loadingStatus, setLoadingStatus] = useState(false);
  const [loadingResult, setLoadingResult] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const pollingRef = useRef<number | null>(null);

  const loadHistory = async () => {
    setHistoryLoading(true);

    try {
      const data = await requestJSON<DownloadListResponse>(`${API_BASE}/downloads`);
      setHistory(data.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载历史记录失败");
    } finally {
      setHistoryLoading(false);
    }
  };

  useEffect(() => {
    void loadHistory();
  }, []);

  useEffect(() => {
    if (!job?.jobId) {
      return;
    }

    if (status?.status === "completed" || status?.status === "failed") {
      if (pollingRef.current) {
        window.clearInterval(pollingRef.current);
        pollingRef.current = null;
      }

      if (status.status === "completed") {
        void handleResult(job.jobId);
        void loadHistory();
      }

      return;
    }

    pollingRef.current = window.setInterval(() => {
      void handleStatus(job.jobId, true);
    }, 3000);

    return () => {
      if (pollingRef.current) {
        window.clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    };
  }, [job?.jobId, status?.status]);

  const handleProbe = async () => {
    setLoadingProbe(true);
    setError(null);
    setNotice(null);

    try {
      const data = await requestJSON<ProbeResponse>(`${API_BASE}/videos/probe`, {
        method: "POST",
        body: JSON.stringify({ url }),
      });
      setProbe(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "探测失败");
    } finally {
      setLoadingProbe(false);
    }
  };

  const handleCreate = async () => {
    setLoadingCreate(true);
    setError(null);
    setNotice(null);

    try {
      if (!probe || probe.videoId === "") {
        const probeData = await requestJSON<ProbeResponse>(`${API_BASE}/videos/probe`, {
          method: "POST",
          body: JSON.stringify({ url }),
        });
        setProbe(probeData);
      }

      const data = await requestJSON<CreateDownloadResponse>(`${API_BASE}/downloads`, {
        method: "POST",
        body: JSON.stringify({ url, outputFormat: format }),
      });
      setJob(data);
      setStatus({
        jobId: data.jobId,
        status: data.status,
        progress: 0,
        step: "queued",
      });
      setResult(null);
      setNotice("任务已创建，系统正在后台下载并自动刷新状态。");
      void loadHistory();
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建任务失败");
    } finally {
      setLoadingCreate(false);
    }
  };

  const handleResult = async (jobId = job?.jobId) => {
    if (!jobId) {
      return;
    }

    setLoadingResult(true);
    setError(null);

    try {
      const data = await requestJSON<DownloadResultResponse>(`${API_BASE}/downloads/${jobId}/result`);
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "获取结果失败");
    } finally {
      setLoadingResult(false);
    }
  };

  const handleStatus = async (jobId = job?.jobId, silent = false) => {
    if (!jobId) {
      return;
    }

    if (!silent) {
      setLoadingStatus(true);
      setError(null);
    }

    try {
      const data = await requestJSON<DownloadStatusResponse>(`${API_BASE}/downloads/${jobId}`);
      setStatus(data);
    } catch (err) {
      if (!silent) {
        setError(err instanceof Error ? err.message : "查询状态失败");
      }
    } finally {
      if (!silent) {
        setLoadingStatus(false);
      }
    }
  };

  return (
    <main className="page-shell">
      <section className="hero-card">
        <div>
          <p className="eyebrow">Cloudflare Proxy + VPS</p>
          <h1>下载 YouTube 视频，不用再盯着终端。</h1>
          <p className="intro">
            输入链接，系统会自动探测、排队、下载、转码并上传到 R2。你只需要等结果，不需要反复刷新。
          </p>
          <div className="hero-actions">
            {authEnabled ? (
              <>
                <Show when="signed-out">
                  <div className="auth-actions">
                    <SignInButton>
                      <button type="button" className="ghost-button">
                        登录
                      </button>
                    </SignInButton>
                    <SignUpButton>
                      <button type="button" className="ghost-button">
                        注册
                      </button>
                    </SignUpButton>
                  </div>
                </Show>
                <Show when="signed-in">
                  <div className="auth-user">
                    <span>账号已连接</span>
                    <UserButton />
                  </div>
                </Show>
              </>
            ) : (
              <p className="error-box">未检测到 `VITE_CLERK_PUBLISHABLE_KEY`，Clerk 登录入口暂未启用。</p>
            )}
          </div>
          <p className="muted-copy auth-copy">Clerk 登录已接入，当前下载与额度逻辑暂时仍沿用现有后端规则；支付与套餐后续再接。</p>
        </div>

        <div className="hero-meta">
          <div>
            <strong>{history.length}</strong>
            <span>最近任务</span>
          </div>
          <div>
            <strong>{status?.status === "completed" ? "已完成" : "处理中"}</strong>
            <span>当前任务</span>
          </div>
        </div>
      </section>

      <section className="main-grid">
        <section className="panel download-panel">
          <div className="panel-head">
            <div>
              <p className="eyebrow">Step 1</p>
              <h2>发起下载</h2>
            </div>
            <button type="button" className="ghost-button" onClick={() => void loadHistory()}>
              {historyLoading ? "刷新中..." : "刷新历史"}
            </button>
          </div>

          <label className="field">
            <span>YouTube URL</span>
            <input value={url} onChange={(e) => setURL(e.target.value)} placeholder="粘贴视频链接" />
          </label>

          <div className="inline-controls">
            <label className="field compact-field">
              <span>输出格式</span>
              <select value={format} onChange={(e) => setFormat(e.target.value as OutputFormat)}>
                <option value="mp4">MP4</option>
                <option value="source">原始格式</option>
              </select>
            </label>

            <div className="actions">
              <button type="button" onClick={handleProbe} disabled={loadingProbe}>
                {loadingProbe ? "探测中..." : "先探测"}
              </button>
              <button type="button" onClick={handleCreate} disabled={loadingCreate}>
                {loadingCreate ? "创建中..." : "开始下载"}
              </button>
            </div>
          </div>

          {notice ? <p className="notice-box">{notice}</p> : null}
          {error ? <p className="error-box">{error}</p> : null}

          {probe ? (
            <div className="probe-card">
              {probe.thumbnailUrl ? <img src={probe.thumbnailUrl} alt={probe.title} /> : null}
              <div>
                <p className="card-label">探测结果</p>
                <h3>{probe.title}</h3>
                <p>
                  视频 ID：{probe.videoId} · 时长：{probe.durationSec}s · 可用格式：
                  {probe.allowedFormats.join(", ")}
                </p>
              </div>
            </div>
          ) : null}
        </section>

        <section className="side-column">
          <article className="result-card current-card">
            <div className="panel-head">
              <div>
                <p className="eyebrow">Step 2</p>
                <h2>当前任务</h2>
              </div>
              <button
                type="button"
                className="ghost-button"
                onClick={() => void handleStatus()}
                disabled={!job?.jobId || loadingStatus}
              >
                {loadingStatus ? "刷新中..." : "立即刷新"}
              </button>
            </div>

            {job ? (
              <dl className="stats-grid">
                <div>
                  <dt>任务 ID</dt>
                  <dd>{job.jobId}</dd>
                </div>
                <div>
                  <dt>状态</dt>
                  <dd>{status?.status ?? job.status}</dd>
                </div>
                <div>
                  <dt>阶段</dt>
                  <dd>{status?.step ?? "queued"}</dd>
                </div>
                <div>
                  <dt>进度</dt>
                  <dd>{status?.progress ?? 0}%</dd>
                </div>
              </dl>
            ) : (
              <p className="muted-copy">还没有任务。创建任务后，这里会自动显示后台进度。</p>
            )}

            {result ? (
              <div className="result-box">
                <p className="card-label">下载结果</p>
                <h3>{result.fileName}</h3>
                <p>{(result.fileSize / 1024 / 1024).toFixed(1)} MB</p>
                <div className="actions">
                  <a className="download-link" href={result.downloadUrl} target="_blank" rel="noreferrer">
                    打开下载链接
                  </a>
                  <button type="button" className="ghost-button" onClick={() => void handleResult()}>
                    {loadingResult ? "获取中..." : "刷新结果"}
                  </button>
                </div>
              </div>
            ) : null}
          </article>

          <article className="result-card history-card">
            <div className="panel-head">
              <div>
                <p className="eyebrow">History</p>
                <h2>最近任务</h2>
              </div>
              <span className="history-count">{history.length}</span>
            </div>

            {history.length > 0 ? (
              <ul className="history-list">
                {history.map((item) => (
                  <li key={item.jobId}>
                    <button
                      type="button"
                      className="history-item"
                      onClick={() => {
                        setJob({ jobId: item.jobId, status: item.status });
                        setStatus({
                          jobId: item.jobId,
                          status: item.status,
                          progress: item.status === "completed" ? 100 : 0,
                          step: item.status === "completed" ? "done" : item.status,
                        });
                        setResult(null);
                        setNotice(null);
                        setError(null);
                        if (item.status === "completed") {
                          void handleResult(item.jobId);
                        }
                      }}
                    >
                      {item.thumbnailUrl ? <img src={item.thumbnailUrl} alt={item.title} /> : <span className="thumb-fallback">YT</span>}
                      <span>
                        <strong>{item.title || item.jobId}</strong>
                        <small>
                          {item.status} · {item.outputFormat} · {new Date(item.createdAt).toLocaleString()}
                        </small>
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="muted-copy">还没有历史任务。你第一次完成下载后，这里会保留最近记录。</p>
            )}
          </article>
        </section>
      </section>
    </main>
  );
}
