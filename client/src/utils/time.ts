export const calcUtcHHMM = (time: string, add: number) => {
  const [h, m] = time.split(':');
  const tempH = Number(h);
  const hour = tempH + add;

  if (hour > 24) {
    if (hour - 24 < 10) {
      return '0' + (hour - 24) + ':' + m;
    }

    return hour + ':' + m;
  }
  return hour + ':' + m;
};

export const getUtcHHMM = () => new Date().toUTCString().substring(17, 22);

export const formatTime = (time: string): string => time.slice(0, 10).replaceAll('-', '/');
