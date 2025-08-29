
export interface AppConfig {
  challengeUrl: string;
}

export interface FckBotType {
  config: AppConfig;
  setConfig(config: Partial<AppConfig>): void;
  fetch?: (url: string | URL, options?: RequestInit) => Promise<Response>;
}

let config: AppConfig = {
  challengeUrl: '',
};

export const FckBot: FckBotType = {
  config,
  setConfig(newConfig) {
    config = { ...config, ...newConfig };
    FckBot.config = config;
  }
};

(window as any).FckBot = FckBot;
