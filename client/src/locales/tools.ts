let prefix: string[] = [];

export function getI18n(origin: any, target: any) {
  let i = 0;

  for (let key in origin) {
    if (typeof origin[key] === 'string') {
      target[[...prefix, key].join('.')] = origin[key];
    } else {
      prefix.push(key);
      getI18n(origin[key], target);
      prefix.pop();
    }
    i++;
  }
}

export const get$t = (intl: any) => (id: string, params: object = {}) => intl.formatMessage({ id }, params);

export const getTimeZone = () => (0 - new Date().getTimezoneOffset()) / 60 + 'h';

export const getTime = (date: string | number): [number, number, number] => {
  const d = new Date(date);
  return [d.getFullYear(), d.getMonth() + 1, d.getDate()];
};

export const getDurationDays = (d1: string | number, d2: string | number): string => {
  const d1_ = new Date(d1);
  const d2_ = new Date(d2);
  return (Math.abs(d2_.getTime() - d1_.getTime()) / 1000 / 60 / 60 / 24).toFixed(0);
};
