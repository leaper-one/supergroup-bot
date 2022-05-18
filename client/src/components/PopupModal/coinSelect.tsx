import React, { useEffect, useState } from 'react';
import { Modal } from 'antd-mobile';
import styles from './coinSelect.less';
import { ApiGetAssetBySymbol, ApiGetMyAssets, ApiGetTop100, IAsset } from '@/apis/asset';
import { ApiGetAdminAndGuest, IUser } from '@/apis/user';
import { GlobalData } from '@/stores/store';

type CoinProps = IAsset | undefined;

interface Props {
  select: (asset: CoinProps) => void;
  closeWithAnimation?: number;
  active?: CoinProps;
  myAsset?: boolean;
  avoid?: string[];
  assetList?: IAsset[];
  $t: (key: string) => string;
}
interface IPopCoinModalProps {
  coinModal: boolean;
  setCoinModal: (v: boolean) => void;
  activeCoin: IAsset | undefined;
  setActiveCoin: (a: IAsset | undefined) => void;
  assetList?: IAsset[];
  $t: (key: string) => string;
}

export const PopCoinModal = (props: IPopCoinModalProps) => {
  return (
    <Modal popup animationType="slide-up" visible={props.coinModal} onClose={() => props.setCoinModal(false)}>
      <CoinModal
        active={props.activeCoin}
        select={(asset) => {
          props.setActiveCoin(asset);
          props.setCoinModal(false);
        }}
        assetList={props.assetList}
        $t={props.$t}
      />
    </Modal>
  );
};
let timer: any = null;
export const CoinModal = (props: Props) => {
  const { select, active, avoid, myAsset } = props;
  const [assetList, setAssetList] = useState<IAsset[]>([]);
  const [search, setSearch] = useState('');

  useEffect(() => {
    if (myAsset) {
      ApiGetMyAssets().then((list) =>
        setAssetList(
          list
            .filter((item) => Number(item.balance) * Number(item.price_usd) > 1)
            .filter((item) => item.name?.toLowerCase().includes(search.toLowerCase()) || item.symbol?.toLowerCase().includes(search.toLowerCase())),
        ),
      );
      return;
    }
    if (search) {
      if (!props.assetList) {
        clearTimeout(timer);
        timer = setTimeout(() => {
          ApiGetAssetBySymbol(search).then((list) => {
            setAssetList(list);
            timer = null;
          });
        }, 500);
      }
    } else setAssetInit();
  }, [search]);

  const setAssetInit = async () => {
    if (myAsset) {
      const list = await ApiGetMyAssets();
      setAssetList(list.filter((item) => Number(item.price_usd) > 0));
    } else if (!props.assetList) {
      let list = await ApiGetTop100();
      if (avoid) list = list.filter((item: IAsset) => !avoid.includes(item.asset_id!));
      setAssetList(list.slice(0, 20));
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.search + ' ' + 'flex'}>
        <img src={require('@/assets/img/svg/search.svg')} alt="" />
        <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Name, Symbol" type="text" />
        <span onClick={() => select(active)}>{props.$t('action.cancel')}</span>
      </div>
      <ul className={styles.list}>
        {(props.assetList || assetList).map((asset, idx) => (
          <li key={asset.asset_id} onClick={() => select(asset)}>
            <img src={asset.icon_url} alt="" />
            <p>{asset.name}</p>
            <i>{myAsset ? `${asset.balance} ${asset.symbol}` : asset.symbol}</i>
            <img className={(active && active.asset_id === asset.asset_id && styles.selected) + ' ' + styles.select} src={require('@/assets/img/svg/select.svg')} alt="" />
          </li>
        ))}
      </ul>
    </div>
  );
};

interface IPopAdminAndGuestModalProps {
  userModal: boolean;
  setUserModal: (v: boolean) => void;
  activeUser: IUser | undefined;
  setActiveUser: (a: IUser | undefined) => void;
  $t: any;
}

export const PopAdminAndGuestModal = (props: IPopAdminAndGuestModalProps) => {
  return (
    <Modal popup animationType="slide-up" visible={props.userModal} onClose={() => props.setUserModal(false)}>
      <UserModal
        active={props.activeUser}
        select={(asset) => {
          props.setActiveUser(asset);
          props.setUserModal(false);
        }}
        $t={props.$t}
      />
    </Modal>
  );
};

interface IUserModalProps {
  active: IUser | undefined;
  select: (a: IUser | undefined) => void;
  $t: any;
}

const userStatusMap = { 8: 'guest', 9: 'admin' };
export const UserModal = (props: IUserModalProps) => {
  const { select, active } = props;
  const [userList, setUserList] = useState<IUser[]>([]);
  const [search, setSearch] = useState('');

  useEffect(() => {
    if (search)
      setUserList(
        GlobalData.adminAndGuests.filter((item: IUser) => item.full_name?.toLowerCase().includes(search.toLowerCase()) || item.identity_number?.toLowerCase().includes(search.toLowerCase())),
      );
    else setUserInit();
  }, [search]);

  const setUserInit = () => {
    ApiGetAdminAndGuest().then(setUserList);
  };

  return (
    <div className={styles.container}>
      <div className={styles.search + ' ' + 'flex'}>
        <img src={require('@/assets/img/svg/search.svg')} alt="" />
        <input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Mixin ID, Name" type="text" />
        <span onClick={() => select(active)}>{props.$t('action.cancel')}</span>
      </div>
      <ul className={styles.list}>
        {userList.map((user, idx) => (
          <li key={user.user_id} onClick={() => select(user)}>
            <img src={user.avatar_url} alt="" />
            <div className={styles.userName}>
              <span>{user.full_name}</span>
              <b>{props.$t(`member.status${user.status}`)}</b>
            </div>
            <i>{user.identity_number}</i>
            <img className={`${active && active.user_id === user.user_id && styles.selected} ${styles.select}`} src={require('@/assets/img/svg/select.svg')} alt="" />
          </li>
        ))}
      </ul>
    </div>
  );
};
