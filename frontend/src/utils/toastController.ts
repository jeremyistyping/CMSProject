/**
 * Toast Controller Utility
 * Prevents toast spam and manages notifications
 */

class ToastController {
  private lastToastTimes: { [key: string]: number } = {};
  private activeToasts: Set<string> = new Set();

  /**
   * Show a throttled toast to prevent spam
   */
  showThrottledToast(
    key: string,
    toastFunction: (config: any) => void,
    toastConfig: any,
    throttleMs: number = 5000
  ): boolean {
    const now = Date.now();
    const lastTime = this.lastToastTimes[key] || 0;

    if (now - lastTime > throttleMs && !this.activeToasts.has(key)) {
      this.lastToastTimes[key] = now;
      this.activeToasts.add(key);

      // Remove from active set after toast duration
      const duration = toastConfig.duration || 5000;
      setTimeout(() => {
        this.activeToasts.delete(key);
      }, duration);

      toastFunction(toastConfig);
      return true;
    }

    return false;
  }

  /**
   * Clear all toast throttles
   */
  clearAll() {
    this.lastToastTimes = {};
    this.activeToasts.clear();
  }

  /**
   * Get throttle status for debugging
   */
  getStatus() {
    return {
      lastToastTimes: { ...this.lastToastTimes },
      activeToasts: Array.from(this.activeToasts),
      now: Date.now()
    };
  }
}

export const toastController = new ToastController();
export default toastController;