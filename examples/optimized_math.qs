// Optimized Math Program with Inline Assembly
// Demonstrates using inline assembly for performance-critical operations

export function entry(ctx: InstructionContext): i64 {
    let x: i64 = 42;
    let y: i64 = 10;
    
    // Use inline assembly for optimized multiplication by power of 2
    let doubled: i64;
    __asm__ {
        LOAD x
        DUP
        ADD
        STORE doubled
    }
    
    // Use inline assembly for optimized division by power of 2
    let halved: i64;
    __asm__ {
        LOAD y
        PUSH i64 2
        DIV
        STORE halved
    }
    
    // Complex calculation using both high-level and assembly
    let result: i64 = doubled + halved;
    
    // Use assembly for bit manipulation
    let shifted: i64;
    __asm__ {
        LOAD result
        PUSH i64 2
        SHL
        STORE shifted
    }
    
    return shifted;
}

// Function demonstrating assembly for stack manipulation
export function stackDemo(a: i64, b: i64, c: i64): i64 {
    let result: i64;
    
    __asm__ {
        // Load all three parameters
        LOAD a
        LOAD b
        LOAD c
        
        // Stack now has: [a, b, c] (c on top)
        // Compute: (a + b) * c
        
        // Swap to get b on top
        SWAP
        // Stack: [a, c, b]
        
        // Swap again to get a on top
        SWAP
        // Stack: [c, a, b]
        
        // Add a and b
        ADD
        // Stack: [c, (a+b)]
        
        // Multiply by c
        MUL
        // Stack: [(a+b)*c]
        
        STORE result
    }
    
    return result;
}

// Function demonstrating conditional logic in assembly
export function maxValue(a: i64, b: i64): i64 {
    let result: i64;
    
    __asm__ {
        LOAD a
        LOAD b
        
        // Duplicate b for comparison
        DUP
        // Stack: [a, b, b]
        
        // Load a again for comparison
        LOAD a
        // Stack: [a, b, b, a]
        
        // Compare: is b > a?
        GT
        // Stack: [a, b, (b>a)]
        
        // If b > a, jump to use_b
        JMPIF use_b
        
        // Otherwise use a
        POP
        STORE result
        JMP end
        
        use_b:
        SWAP
        POP
        STORE result
        
        end:
    }
    
    return result;
}
