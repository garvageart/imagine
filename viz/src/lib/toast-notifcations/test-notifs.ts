import { toastState } from "./notif-state.svelte";

export function addTestNotifications() {
    toastState.addToast({
        type: "info",
        title: "Welcome to Imagine!",
        message: "This is an **info** notification with a *title*. Check out the [documentation](https://imagine.com/docs).",
        timeout: 0, // Infinite
        actions: [
            { label: "Dismiss", onClick: () => toastState.dismissToast(0) }
        ]
    });

    toastState.addToast({
        type: "success",
        title: "Image Uploaded",
        message: "Your image 'my_photo.jpg' has been successfully uploaded. Go to [My Photos](/photos) to see it.",
        timeout: 0, // Infinite
        actions: [
            { label: "View Photo", onClick: () => alert("Viewing photo!") },
            { label: "Undo", onClick: () => alert("Undo upload!") }
        ]
    });

    toastState.addToast({
        type: "warning",
        message: "Disk space is running low. Please clear some space. More info at https://imagine.com/storage",
        timeout: 0, // Infinite
    });

    toastState.addToast({
        type: "error",
        title: "Server Error (500)",
        message: "Failed to connect to the server. Please try again later.",
        timeout: 0, // Infinite
    });

    toastState.addToast({
        type: "info",
        message: "This is a plain info message without a title or actions.",
        timeout: 0, // Infinite
    });
}
