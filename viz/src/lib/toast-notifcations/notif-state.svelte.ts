export interface ToastAction {
  label: string;
  onClick: () => void;
}

export interface Toast {
  id: number;
  message: string;
  title?: string;
  actions?: ToastAction[];
  dismissible?: boolean;
  timeout?: number;
  type?: "success" | "info" | "warning" | "error";
}

class ToastState {
  toasts = $state<Toast[]>([]);

  dismissToast = (id: number) => {
    this.toasts = this.toasts.filter((toast) => toast.id !== id);
  };

  /**
   * Original code: https://svelte.dev/repl/0091c8b604b74ed88bb7b6d174504f50?version=3.35.0
   *
   * Default timeout is 3000ms (3 seconds)
   */
  addToast = (toast: Partial<Omit<Toast, "id">> = {}) => {
    // Create a unique ID so we can easily find/remove it
    // if it is dismissible/has a timeout.
    const id = Math.floor(Math.random() * 10000);

    // Setup some sensible defaults for a toast and merge with passed values
    const defaultToast = {
      dismissible: true,
      timeout: 3000,
      type: "info" as const,
      message: "No message to display",
    };

    const mergedToast = { ...defaultToast, ...toast, id };

    // Push the toast to the top of the list of toasts
    this.toasts.unshift(mergedToast);

    // If toast is dismissible and has a timeout, dismiss it after "timeout" amount of time.
    if (mergedToast.timeout && mergedToast.timeout > 0) {
      setTimeout(() => this.dismissToast(id), mergedToast.timeout);
    }
  };
}

export const toastState = new ToastState();
