import { FormEvent, useMemo, useState } from "react";
import { CreateLinkResponse, createShortLink } from "../api/client";
import ResultCard from "./ResultCard";

export default function ShortenForm() {
  const [url, setURL] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<CreateLinkResponse | null>(null);

  const disabled = useMemo(() => loading || url.trim().length === 0, [loading, url]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!url.trim()) {
      setError("Введите ссылку, которую нужно сократить");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const payload = await createShortLink({ url: url.trim() });
      setResult(payload);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Не удалось сократить ссылку";
      setError(message);
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <section className="card">
      <div className="card-header">
        <h2>Сократите ссылку</h2>
        <p>Сервис вернет аккуратный short URL, который можно сразу отправлять</p>
      </div>

      <form onSubmit={onSubmit} className="form">
        <label htmlFor="shorten-url">URL</label>
        <div className="input-row">
          <input
            id="shorten-url"
            value={url}
            onChange={(e) => setURL(e.target.value)}
            placeholder="Введите ссылку, которую нужно сократить"
            autoComplete="off"
          />
          <button className="button button-accent" type="submit" disabled={disabled}>
            {loading ? "Сокращаем..." : "Сократить"}
          </button>
        </div>
      </form>

      <div className="result-slot">
        {error ? <div className="inline-error">{error}</div> : null}

        {result ? (
          <ResultCard
            title="Готово"
            value={result.short_url}
            subtitle={url.trim()}
          />
        ) : null}
      </div>
    </section>
  );
}
