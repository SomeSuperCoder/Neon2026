// Example: File Operations
// Level: 4
// Features: getFile, len function, bytes type
// Description: Demonstrates reading file data from the FileStore

export function entry(): i64 {
    // Get file data using index 0
    let data: bytes = getFile(0);
    
    // Return the length of the data
    return len(data);
}
