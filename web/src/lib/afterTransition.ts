/**
 * Calls `done` once `el`'s own transition ends, or straight away if it has
 * none. A timeout stands in for a `transitionend` that never comes, e.g. when
 * the transition is interrupted. Transitions inside `el` are ignored.
 */
export function afterTransition(el: HTMLElement, done: () => void) {
  const seconds = parseFloat(getComputedStyle(el).transitionDuration) || 0;
  if (seconds === 0) {
    done();
    return;
  }
  let hasFinished = false;
  const finish = () => {
    if (hasFinished) return;
    hasFinished = true;
    el.removeEventListener("transitionend", onEnd);
    done();
  };
  const onEnd = (event: Event) => {
    if (event.target === el) finish();
  };
  el.addEventListener("transitionend", onEnd);
  setTimeout(finish, seconds * 1000 + 100);
}
