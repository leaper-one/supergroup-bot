import { request } from '@@/plugin-request/request';
import { apis } from '@/apis/http';
import { getGroupID } from '@/apis/group';

export interface IPrsdiggArticle {
  uuid: string;
  title: string;
  price_usd: number;
  price: number;
  currency: string;
  intro: string;
  original_url: string;
  partial_content: string | null;
  state: string;
  orders_count: number;
  comments_count: number;
  upvotes_count: number;
  downvotes_count: number;
  tag_names: string[];
  author: {
    name: string;
    avatar: string;
  };
  created_at: string;
  updated_at: string;

  asset_url?: string;
}

export const ApiGetPrsdigg = (query: string, offset = 0): Promise<IPrsdiggArticle[]> => request(`https://prsdigg.com/api/articles?limit=20&&query=${query}&offset=${offset}`);

interface IArticleParams {
  group_id: string;
  title: string;
  link: string;
  key_word: string;
}

export const ApiPostArticle = (params: IArticleParams) => apis.post(`/article`, params);

export interface IArticle {
  article_id: string;
  asset_id: string;
  avatar_url: string;
  created_at: string;
  full_name: string;
  group_id: string;
  icon_url: string;
  key_word: string;
  link: string;
  price_usd: string;
  score: string;
  status: string;
  symbol: string;
  title: string;
  user_id: string;
  comment?: string;

  platform?: string;
  rank?: number;

  amount?: string;
  amount_usd?: string;

  pass?: boolean;
  rate?: string;
  identity_number?: string;
}

export const ApiGetArticle = (status: string): Promise<IArticle[]> => apis.get(`/article/${getGroupID()}/${status}`);

interface IPutArticleParams {
  article_id: string;
  user_id: string;
  group_id: string;
  score: string;
  comment: string;
}

export const ApiPutArticle = (params: IPutArticleParams) => apis.put(`/article/score`, params);
export const ApiPutArticlePass = (article_id: string) => apis.put(`/article/pass`, { article_id });

export const ApiGetArticleCrawlList = () => apis.get(`/article/crawl/${getGroupID()}`);

interface IPutArticleCrawlParams {
  article_id: string;
  created_at: string;
  link: string;
  platform: string;
  rank: number;
}

export const ApiPutArticleCrawl = (params: IPutArticleCrawlParams) => apis.put(`/article/crawl`, params);

export const ApiGetArticleCrawlListByID = (id: string) => apis.get(`/article/crawl/${getGroupID()}/${id}`);
