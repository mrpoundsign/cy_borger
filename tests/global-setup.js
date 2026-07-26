module.exports = async () => {
    console.log("Sleeping for 5 seconds to let server finish booting...");
    await new Promise(resolve => setTimeout(resolve, 5000));
};
