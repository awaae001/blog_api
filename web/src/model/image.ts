export interface Image {
  id: number;
  name: string;
  url: string;
  local_path: string;
  is_local: number;
  status: string;
}

export interface ImageListParams {
  page?: number;
  page_size?: number;
  status?: string;
  search?: string;
 }

export interface PaginatedImages {
  items: Image[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateImagePayload {
  name: string;
  url: string;
  local_path?: string;
  is_local?: number;
}

export interface UpdateImagePayload {
  name?: string;
  url?: string;
  status?: string;
}