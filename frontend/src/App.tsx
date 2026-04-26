import ResolveForm from "./components/ResolveForm";
import ShortenForm from "./components/ShortenForm";

export default function App() {
  return (
    <div className="page">
      <div className="bg-orb bg-orb-left" />
      <div className="bg-orb bg-orb-right" />

      <header className="topbar container">
        <div className="brand-row">
          <div className="logo-dot" />
          <span className="brand">LinkLite</span>
        </div>
      </header>

      <section className="hero container">
        <div className="hero-title-row">
          <span className="hero-icon" aria-hidden="true">
            <svg viewBox="0 0 32 32" fill="none">
              <rect x="2.5" y="2.5" width="27" height="27" rx="9.5" stroke="currentColor" strokeOpacity="0.2" />
              <path
                d="M12.2 18.7l-1.65 1.65a3.2 3.2 0 104.53 4.53l1.64-1.65M19.8 13.3l1.65-1.65a3.2 3.2 0 10-4.53-4.53l-1.64 1.65M12.4 19.6h7.2"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </span>
          <h1 className="hero-title">Короткие ссылки без лишнего шума</h1>
        </div>
        <p>
          Вставьте URL, получите аккуратную короткую ссылку и сразу поделитесь ей.
        </p>
      </section>

      <main className="container forms-grid">
        <ShortenForm />
        <ResolveForm />
      </main>
    </div>
  );
}
