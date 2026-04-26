import { useState } from "react";

interface ResultCardProps {
  title: string;
  value: string;
  subtitle?: string;
  note?: string;
  openLabel?: string;
}

async function copyToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }

    const area = document.createElement("textarea");
    area.value = text;
    area.setAttribute("readonly", "");
    area.style.position = "fixed";
    area.style.opacity = "0";
    document.body.appendChild(area);
    area.focus();
    area.select();
    const success = document.execCommand("copy");
    document.body.removeChild(area);
    return success;
  } catch {
    return false;
  }
}

export default function ResultCard({ title, value, subtitle, note, openLabel = "Открыть" }: ResultCardProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    const ok = await copyToClipboard(value);
    setCopied(ok);
    if (ok) {
      setTimeout(() => setCopied(false), 1500);
    }
  };

  return (
    <article className="result-card" aria-live="polite">
      <h3>{title}</h3>
      <a href={value} target="_blank" rel="noreferrer" className="result-link">
        {value}
      </a>
      {subtitle ? <p className="result-subtitle">{subtitle}</p> : null}
      {note ? <p className="result-note">{note}</p> : null}
      <div className="result-actions">
        <button type="button" className="button button-secondary" onClick={handleCopy}>
          {copied ? "Скопировано" : "Скопировать"}
        </button>
        <a className="button button-primary" href={value} target="_blank" rel="noreferrer">
          {openLabel}
        </a>
      </div>
    </article>
  );
}
