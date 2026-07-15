// material design icon paths (Apache-2.0), inlined so the binary stays
// self-contained — no icon font, no CDN.

function Icon({ d, label, onClick }: { d: string; label: string; onClick: () => void }) {
  return (
    <button className="icon-button" title={label} aria-label={label} onClick={onClick}>
      <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
        <path d={d} />
      </svg>
    </button>
  );
}

export function EditIcon({ onClick }: { onClick: () => void }) {
  return (
    <Icon
      label="edit raw content"
      onClick={onClick}
      d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34a.9959.9959 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"
    />
  );
}

export function CloseIcon({ onClick }: { onClick: () => void }) {
  return (
    <Icon
      label="close"
      onClick={onClick}
      d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
    />
  );
}
