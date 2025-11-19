export interface ITunesMusic {
  wrapperType: string;
  artworkUrl100: string;
  artistName: string;
  collectionCensoredName: string;
  trackViewUrl: string;
  collectionId: number;
  // Optional fields from backend for richer cards
  pressure?: string;
  riskName?: string;
  code?: string;
  description?: string;
}
export interface ITunesResult {
  resultCount: number;
  results: ITunesMusic[];
}

const API_BASE = (import.meta.env.VITE_API_BASE as string) || '';
// Sensible defaults for dev if .env not provided
const MINIO_PUBLIC_BASE = (import.meta.env.VITE_MINIO_PUBLIC_BASE as string) || 'http://localhost:9000';
const MINIO_BUCKET = (import.meta.env.VITE_MINIO_BUCKET as string) || 'images';

export function stageToITunes(stage: any): ITunesMusic {
  const id = stage.ID ?? stage.id ?? stage.stage_id ?? stage.service_id;
  const title = stage.Title ?? stage.title ?? stage.name ?? stage.label ?? 'Без названия';
  const description = stage.Description ?? stage.description ?? stage.desc ?? stage.details ?? '';
  const risk = stage.RiskName ?? stage.RiskClass ?? stage.risk_class ?? stage.riskClass ?? '';
  const code = stage.Code ?? stage.code ?? '';
  const pressure = stage.Pressure ?? stage.pressure ?? '';

  const imageUrlFull = stage.ImageUrl ?? stage.image_url ?? stage.imageUrl ?? '';
  const imageKey = stage.ImageKey ?? stage.image_key ?? stage.imageKey ?? stage.image ?? '';
  const artwork = imageUrlFull
    ? String(imageUrlFull)
    : (imageKey
        ? `${MINIO_PUBLIC_BASE.replace(/\/$/, '')}/${MINIO_BUCKET}/${imageKey}`
        : '');

  return {
    wrapperType: 'stage',
    artworkUrl100: artwork,
    artistName: String(risk || description),
    collectionCensoredName: String(title),
    trackViewUrl: `/albums/${id ?? ''}`,
    collectionId: id,
    pressure: pressure ? String(pressure) : undefined,
    riskName: risk ? String(risk) : undefined,
    code: code ? String(code) : undefined,
    description: description ? String(description) : undefined,
  } as ITunesMusic;
}

export const getMusicByName = async (name = ''): Promise<ITunesResult> => {
  return getStages({ query: name });
};

export const getAlbumById = async (id: number | string): Promise<ITunesResult> => {
  const endpoint = API_BASE ? `${API_BASE}/api/stages/${id}` : `/api/stages/${id}`;
  const res = await fetch(endpoint, { headers: { Accept: 'application/json' } });
  if (!res.ok) {
    throw new Error(`Backend error: ${res.status}`);
  }
  const stage = await res.json();
  const item = stageToITunes(stage);
  return { resultCount: 1, results: [item] };
};

export type StageRaw = Record<string, any>;

export const getStageByIdRaw = async (id: number | string): Promise<StageRaw> => {
  const endpoint = API_BASE ? `${API_BASE}/api/stages/${id}` : `/api/stages/${id}`;
  const res = await fetch(endpoint, { headers: { Accept: 'application/json' } });
  if (!res.ok) {
    throw new Error(`Backend error: ${res.status}`);
  }
  return res.json();
};

export interface StageFilters {
  query?: string;
  // Pressure filters (backend params)
  sys_from?: number;
  sys_to?: number;
  dia_from?: number;
  dia_to?: number;
}

export const getStages = async (filters: StageFilters): Promise<ITunesResult> => {
  const params = new URLSearchParams();
  if (filters.query) params.set('query', filters.query);
  if (typeof filters.sys_from === 'number') {
    params.set('sys_from', String(filters.sys_from));
    params.set('sysFrom', String(filters.sys_from));
  }
  if (typeof filters.sys_to === 'number') {
    params.set('sys_to', String(filters.sys_to));
    params.set('sysTo', String(filters.sys_to));
  }
  if (typeof filters.dia_from === 'number') {
    params.set('dia_from', String(filters.dia_from));
    params.set('diaFrom', String(filters.dia_from));
  }
  if (typeof filters.dia_to === 'number') {
    params.set('dia_to', String(filters.dia_to));
    params.set('diaTo', String(filters.dia_to));
  }

  const qs = params.toString();
  const endpoint = API_BASE
    ? `${API_BASE}/api/stages${qs ? `?${qs}` : ''}`
    : `/api/stages${qs ? `?${qs}` : ''}`;

  const res = await fetch(endpoint, { headers: { Accept: 'application/json' } });
  if (!res.ok) {
    throw new Error(`Backend error: ${res.status}`);
  }

  const data = await res.json();
  let stages: any[] = Array.isArray(data) ? data : data.items || data.results || [];

  // Client-side fallback filtering in case backend ignores filters
  const q = (filters.query || '').trim().toLowerCase();
  const sFrom = typeof filters.sys_from === 'number' ? filters.sys_from : undefined;
  const sTo = typeof filters.sys_to === 'number' ? filters.sys_to : undefined;
  const dFrom = typeof filters.dia_from === 'number' ? filters.dia_from : undefined;
  const dTo = typeof filters.dia_to === 'number' ? filters.dia_to : undefined;

  if (q || sFrom !== undefined || sTo !== undefined || dFrom !== undefined || dTo !== undefined) {
    const num = (v: any): number | undefined => {
      const n = Number(v);
      return Number.isFinite(n) ? n : undefined;
    };

    const get = (o: any, keys: string[]): any => {
      for (const k of keys) if (o?.[k] !== undefined && o?.[k] !== null) return o[k];
      return undefined;
    };

    stages = stages.filter((st: any) => {
      const title = String(get(st, ['Title','title','name','label']) || '').toLowerCase();
      if (q && !title.includes(q)) return false;

      const sf = num(get(st, ['SysFrom','sys_from','sysFrom']));
      const stp = num(get(st, ['SysTo','sys_to','sysTo']));
      const df = num(get(st, ['DiaFrom','dia_from','diaFrom']));
      const dt = num(get(st, ['DiaTo','dia_to','diaTo']));

      // Overlap checks: stage range intersects requested range (if provided)
      const sysOk = (sFrom === undefined && sTo === undefined) || (
        sf !== undefined && stp !== undefined &&
        (sTo === undefined || sf <= sTo) && (sFrom === undefined || stp >= sFrom)
      );
      const diaOk = (dFrom === undefined && dTo === undefined) || (
        df !== undefined && dt !== undefined &&
        (dTo === undefined || df <= dTo) && (dFrom === undefined || dt >= dFrom)
      );
      return sysOk && diaOk;
    });
  }

  const results = stages.map(stageToITunes);
  return { resultCount: results.length, results };
};