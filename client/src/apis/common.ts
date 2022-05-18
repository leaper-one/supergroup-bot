import { apis } from '@/apis/http';

interface IUploadFileRes {
  view_url: string;
}

export const ApiUploadFile = (file: File): Promise<IUploadFileRes> => {
  const data = new FormData();
  data.append('file', file);
  return apis.post(`/upload`, data);
};
