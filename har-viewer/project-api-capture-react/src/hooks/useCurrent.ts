import { useRef, useEffect } from "react";

export function useCurrent<T>(value: T) {
  const ref = useRef(value);
  useEffect(() => {
    ref.current = value;
  });
  return ref;
}
