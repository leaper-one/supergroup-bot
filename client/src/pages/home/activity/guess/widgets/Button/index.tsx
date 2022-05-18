import React, { FC, HTMLProps } from 'react';

import styles from './button.less';

export interface ButtonProps extends HTMLProps<HTMLButtonElement> {
  kind?: 'primary' | 'warning';
  type?: 'button' | 'submit' | 'reset';
}

export const Button: FC<ButtonProps> = ({ children, className, kind = 'primary', ...rest }) => {
  return (
    <button {...rest} className={`${styles.button} ${styles[kind]} ${className}`}>
      {children}
    </button>
  );
};
