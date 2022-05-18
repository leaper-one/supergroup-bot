import { apis, mixinBaseURL } from './http';
import { GlobalData } from '@/stores/store';
import { $get } from '@/stores/localStorage';

export interface IAsset {
  type?: string;
  asset_id?: string;
  chain_id?: string;
  symbol?: string;
  name?: string;
  icon_url?: string;
  price_btc?: string;
  price_usd?: string;
  change_btc?: string;
  change_usd?: string;
  asset_key?: string;
  mixin_id?: string;
  amount?: string;
  balance?: string;
}

export const ApiGetTop100 = async (): Promise<IAsset[]> => {
  if (GlobalData.CoinAssetList) return GlobalData.CoinAssetList;

  let assets: IAsset[] = await apis.get(`${mixinBaseURL}/network/assets/top`);
  assets = assets.map(({ asset_id, symbol, icon_url, name, price_usd }) => ({
    asset_id,
    symbol,
    icon_url,
    name,
    price_usd,
  }));
  GlobalData.CoinAssetList = assets;
  return assets;
};

export const ApiGetAssetByID = (id: string): Promise<IAsset> => apis.get(`${mixinBaseURL}/network/assets/${id}`);

export const ApiGetAssetBySymbol = async (symbol: string): Promise<IAsset[]> => {
  if (!GlobalData.AssetSymbol) GlobalData.AssetSymbol = {};
  if (!GlobalData.AssetSymbol[symbol]) GlobalData.AssetSymbol[symbol] = await apis.get(`${mixinBaseURL}/network/assets/search/${symbol}`);
  return GlobalData.AssetSymbol[symbol];
};

export const ApiGetMyAssets = async (): Promise<IAsset[]> => {
  if (!GlobalData.MyAssetList) {
    const { access_token } = $get('user') || {};
    const assetList: Array<IAsset> = await apis.get(
      `${mixinBaseURL}/assets`,
      {},
      {
        headers: { Authorization: 'Bearer ' + access_token },
      },
    );
    GlobalData.MyAssetList = assetList.filter((item) => item.balance !== '0').sort((a, b) => price(b) - price(a));
  }
  return GlobalData.MyAssetList;
};

const price = (asset: IAsset) => Number(asset.balance) * Number(asset.price_usd);

interface ICheckPaidParams {
  amount: string;
  asset_id: string;
  counter_user_id: string;
  trace_id: string;
}

interface ICheckPaidResp {
  status: string;
}

export const ApiCheckIsPaid = (params: ICheckPaidParams): Promise<ICheckPaidResp> => apis.post(`${mixinBaseURL}/payments`, params);
