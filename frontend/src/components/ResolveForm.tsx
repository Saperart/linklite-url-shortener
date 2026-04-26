import { FormEvent, useMemo, useState } from "react";
import { ResolveLinkResponse, resolveShortCode } from "../api/client";
import ResultCard from "./ResultCard";

function extractCode(input: string): string {
  const trimmed = input.trim();
  if (!trimmed) {
    return "";
  }
  try {
    const parsed = new URL(trimmed);
    return parsed.pathname.replace(/^\/+/, "").split("/")[0] || "";
  } catch {
    return trimmed.replace(/^\/+/, "").split("/")[0] || "";
  }
}

export default function ResolveForm() {
  const [value, setValue] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ResolveLinkResponse | null>(null);

  const disabled = useMemo(() => loading || value.trim().length === 0, [loading, value]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const code = extractCode(value);
    if (!code) {
      setError("Введите короткий код");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const payload = await resolveShortCode(code);
      setResult(payload);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Не удалось получить оригинальную ссылку";
      setError(message);
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="card">
      <div className="card-header">
        <h2>Раскройте short code</h2>
        <p>Вставьте код или короткий URL, чтобы получить оригинал</p>
      </div>

      <form onSubmit={onSubmit} className="form">
        <label htmlFor="resolve-code">Код</label>
        <div className="input-row">
          <input
            id="resolve-code"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder="Введите короткий код"
            autoComplete="off"
          />
          <button className="button button-accent" type="submit" disabled={disabled}>
            {loading ? "Ищем..." : "Найти оригинал"}
          </button>
        </div>
      </form>

      <div className="result-slot">
        {error ? <div className="inline-error">{error}</div> : null}

        {result ? (
          <ResultCard title="Оригинальная ссылка" value={result.original_url} openLabel="Открыть" />
        ) : null}
      </div>
    </section>
  );
}
