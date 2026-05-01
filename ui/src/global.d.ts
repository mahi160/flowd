import type { RawPayload } from './lib/types';

declare global {
  interface Window {
    __FLOWD_DATA__?: RawPayload;
  }
}

export {};
