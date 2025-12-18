export interface SystemConfig {
  system_conf: {
    safe_conf: SafeConfig;
    data_conf: DataConfig;
    crawler_conf: CrawlerConfig;
  };
}

export interface SafeConfig {
  cors_allow_hostlist: string[];
  exclude_paths: string[];
  allow_extension: string[];
}

export interface DataConfig {
  database: DatabaseConfig;
  image: ImageConfig;
  resource: ResourceConfig;
}

export interface DatabaseConfig {
  path: string;
}

export interface ImageConfig {
  path: string;
  conv_to: string;
}

export interface ResourceConfig {
  path: string;
}

export interface CrawlerConfig {
  concurrency: number;
}