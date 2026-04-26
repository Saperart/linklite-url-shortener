export interface CreateLinkRequest {
  url: string;
}

export interface CreateLinkResponse {
  short_url: string;
}

export interface ResolveLinkResponse {
  original_url: string;
}

interface ErrorResponse {
  error?: {
    code?: string;
    message?: string;
  };
}

function normalizeApiError(status: number, payload: ErrorResponse | null): string {
  const backendMessage = payload?.error?.message?.trim();
  if (backendMessage) {
    return backendMessage;
  }
  if (status === 404) {
    return "Код не найден";
  }
  if (status === 400) {
    return "Проверьте корректность данных";
  }
  if (status === 503) {
    return "Сервис временно недоступен";
  }
  return "Произошла ошибка. Попробуйте еще раз";
}

async function safeJson<T>(response: Response): Promise<T | null> {
  try {
    return (await response.json()) as T;
  } catch {
    return null;
  }
}

export async function createShortLink(input: CreateLinkRequest): Promise<CreateLinkResponse> {
  const response = await fetch("/api/v1/links", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input)
  });

  if (!response.ok) {
    const payload = await safeJson<ErrorResponse>(response);
    throw new Error(normalizeApiError(response.status, payload));
  }

  const payload = await safeJson<CreateLinkResponse>(response);
  if (!payload) {
    throw new Error("Пустой ответ сервера");
  }
  return payload;
}

export async function resolveShortCode(code: string): Promise<ResolveLinkResponse> {
  const response = await fetch(`/api/v1/links/${encodeURIComponent(code)}`);
  if (!response.ok) {
    const payload = await safeJson<ErrorResponse>(response);
    throw new Error(normalizeApiError(response.status, payload));
  }

  const payload = await safeJson<ResolveLinkResponse>(response);
  if (!payload) {
    throw new Error("Пустой ответ сервера");
  }
  return payload;
}
