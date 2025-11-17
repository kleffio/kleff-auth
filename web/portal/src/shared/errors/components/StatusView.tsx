import "./StatusView.css";

type StatusViewProps = {
  status: number;
  message: string;
};

export function StatusView({ status, message }: StatusViewProps) {
  return (
    <div className="status-root">
      <div className="status-wrap">
        <div className="status-badge">
          <span className="dot" />
          <span className="label">Error</span>
          <span className="code">{status}</span>
        </div>

        <h1 className="status-number">{status}</h1>

        <p className="status-message">{message}</p>

        <div className="status-actions">
          <button
            onClick={() => window.location.reload()}
            className="button button-primary"
          >
            Retry
          </button>

          <button
            onClick={() => (window.location.href = "/")}
            className="button button-secondary"
          >
            Go Home
          </button>
        </div>

        <p className="status-tip">
          If the issue persists, the backend may be offline.
        </p>
      </div>
    </div>
  );
}
