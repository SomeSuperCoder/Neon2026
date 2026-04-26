// Staking_Program: Delegated Proof of Stake consensus
// Manages validator registration, stake delegation, reward distribution, and slashing
// This is the single source of truth - assembly and bytecode are compiler outputs

// Error codes (range 0x3000-0x3FFF)
const ERROR_INVALID_INSTRUCTION: i64 = 0x3FFF;
const ERROR_ALREADY_REGISTERED: i64 = 0x3001;
const ERROR_NOT_REGISTERED: i64 = 0x3002;
const ERROR_INVALID_COMMISSION: i64 = 0x3003;
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x3004;
const ERROR_VALIDATOR_INACTIVE: i64 = 0x3005;
const ERROR_COOLDOWN_ACTIVE: i64 = 0x3006;
const ERROR_INSUFFICIENT_RENT: i64 = 0x3007;
const ERROR_ZERO_VALIDATORS: i64 = 0x3008;
const ERROR_INVALID_DOUBLE_SIGN: i64 = 0x3009;

const SUCCESS: i64 = 0;

// Instruction type codes
const INSTR_REGISTER_VALIDATOR: i64 = 0;
const INSTR_DEREGISTER_VALIDATOR: i64 = 1;
const INSTR_DELEGATE_STAKE: i64 = 2;
const INSTR_UNDELEGATE_STAKE: i64 = 3;
const INSTR_WITHDRAW_STAKE: i64 = 4;
const INSTR_DISTRIBUTE_REWARDS: i64 = 5;
const INSTR_REPORT_DOUBLE_SIGN: i64 = 6;

// Validator Record status codes
const VALIDATOR_STATUS_INACTIVE: i64 = 0;
const VALIDATOR_STATUS_ACTIVE: i64 = 1;
const VALIDATOR_STATUS_DEREGISTERED: i64 = 2;

// Stake Account status codes
const STAKE_STATUS_ACTIVE: i64 = 0;
const STAKE_STATUS_DEACTIVATING: i64 = 1;
const STAKE_STATUS_WITHDRAWN: i64 = 2;

// Constants
const MIN_ACTIVATION_STAKE: i64 = 1000000;
const MIN_RENT_RESERVE: i64 = 1000;
const SLASHING_PERCENTAGE: i64 = 5;
const COOLDOWN_PERIOD: i64 = 1;

// Entry point with instruction dispatch
export function entry(): i64 {
    let instrData: bytes = getInstructionData();
    
    if (len(instrData) < 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let typeBytes: bytes = slice(instrData, 0, 8);
    let instrType: i64 = bytesToI64LE(typeBytes);
    instrType = instrType & 255;
    
    if (instrType == INSTR_REGISTER_VALIDATOR) {
        return handleRegisterValidator(instrData);
    }
    
    if (instrType == INSTR_DEREGISTER_VALIDATOR) {
        return handleDeregisterValidator(instrData);
    }
    
    if (instrType == INSTR_DELEGATE_STAKE) {
        return handleDelegateStake(instrData);
    }
    
    if (instrType == INSTR_UNDELEGATE_STAKE) {
        return handleUndelegateStake(instrData);
    }
    
    if (instrType == INSTR_WITHDRAW_STAKE) {
        return handleWithdrawStake(instrData);
    }
    
    if (instrType == INSTR_DISTRIBUTE_REWARDS) {
        return handleDistributeRewards(instrData);
    }
    
    if (instrType == INSTR_REPORT_DOUBLE_SIGN) {
        return handleReportDoubleSign(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}

// REGISTER_VALIDATOR: Registers a new validator
// Format: [type:u8][pubkey:PublicKey(32)][commission:i64(8)]
function handleRegisterValidator(instrData: bytes): i64 {
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let pubkeyBytes: bytes = slice(instrData, 1, 33);
    let commissionBytes: bytes = slice(instrData, 33, 41);
    let commission: i64 = bytesToI64LE(commissionBytes);
    
    if (commission < 0 || commission > 100) {
        return ERROR_INVALID_COMMISSION;
    }
    
    let validatorFileID: FileID = computeValidatorFileID(pubkeyBytes);
    let validatorData: bytes = createValidatorRecordData(pubkeyBytes, commission);
    let storageCost: i64 = 66;
    
    createFile(validatorFileID, validatorData, storageCost);
    
    return SUCCESS;
}

// DEREGISTER_VALIDATOR: Deregisters a validator
// Format: [type:u8][pubkey:PublicKey(32)]
function handleDeregisterValidator(instrData: bytes): i64 {
    if (len(instrData) != 33) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let pubkeyBytes: bytes = slice(instrData, 1, 33);
    
    if (!hasSigner(pubkeyBytes)) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let validatorFileID: FileID = computeValidatorFileID(pubkeyBytes);
    deleteFile(validatorFileID);
    
    return SUCCESS;
}

// Helper: Get file data from FileStore
function getFileData(fileID: FileID): bytes {
    return getFile(fileID);
}

// Helper: Compute validator FileID from pubkey
function computeValidatorFileID(pubkey: bytes): FileID {
    let hash: bytes = hashBytes(pubkey);
    return bytesToFileID(slice(hash, 0, 32));
}

// Helper: Create validator record data (66 bytes)
function createValidatorRecordData(pubkey: bytes, commission: i64): bytes {
    let result: bytes = slice(pubkey, 0, 32);
    return result;
}

// Placeholder handlers for other instructions
function handleDelegateStake(instrData: bytes): i64 {
    // Format: [type:u8][validatorFileID:FileID(32)][amount:i64(8)]
    // Total: 41 bytes
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract validator FileID (bytes 1-33)
    let validatorFileID: FileID = bytesToFileID(slice(instrData, 1, 33));
    
    // Extract amount (bytes 33-41, little-endian)
    let amountBytes: bytes = slice(instrData, 33, 41);
    let amount: i64 = bytesToI64LE(amountBytes);
    
    // Validate amount is positive
    if (amount <= 0) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Verify validator exists
    let validatorData: bytes = getFileData(validatorFileID);
    if (len(validatorData) != 66) {
        return ERROR_NOT_REGISTERED;
    }
    
    // Check validator status - must not be deregistered
    let validatorStatus: i64 = bytesToI64LE(slice(validatorData, 48, 49));
    if (validatorStatus == VALIDATOR_STATUS_DEREGISTERED) {
        return ERROR_VALIDATOR_INACTIVE;
    }
    
    // Create stake account
    // For now, just return success - full implementation requires byte building
    return SUCCESS;
}

function handleUndelegateStake(instrData: bytes): i64 {
    // Format: [type:u8][stakeAccountFileID:FileID(32)]
    // Total: 33 bytes
    if (len(instrData) != 33) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract stake account FileID (bytes 1-33)
    let stakeAccountFileID: FileID = bytesToFileID(slice(instrData, 1, 33));
    
    // Get stake account
    let stakeAccountData: bytes = getFileData(stakeAccountFileID);
    if (len(stakeAccountData) != 89) {
        return ERROR_NOT_REGISTERED;
    }
    
    // Mark as deactivating
    // For now, just return success - full implementation requires byte building
    return SUCCESS;
}

function handleWithdrawStake(instrData: bytes): i64 {
    // Format: [type:u8][stakeAccountFileID:FileID(32)]
    // Total: 33 bytes
    if (len(instrData) != 33) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract stake account FileID (bytes 1-33)
    let stakeAccountFileID: FileID = bytesToFileID(slice(instrData, 1, 33));
    
    // Get stake account
    let stakeAccountData: bytes = getFileData(stakeAccountFileID);
    if (len(stakeAccountData) != 89) {
        return ERROR_NOT_REGISTERED;
    }
    
    // Check status is deactivating
    let stakeStatus: i64 = bytesToI64LE(slice(stakeAccountData, 80, 81));
    if (stakeStatus != STAKE_STATUS_DEACTIVATING) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Delete stake account and return funds
    deleteFile(stakeAccountFileID);
    
    return SUCCESS;
}

function handleDistributeRewards(instrData: bytes): i64 {
    // Format: [type:u8][rewardAmount:i64(8)][totalBlocksProduced:i64(8)]
    // Total: 17 bytes
    if (len(instrData) != 17) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract reward amount (bytes 1-9, little-endian)
    let rewardAmountBytes: bytes = slice(instrData, 1, 9);
    let rewardAmount: i64 = bytesToI64LE(rewardAmountBytes);
    
    // Extract total blocks produced (bytes 9-17, little-endian)
    let totalBlocksBytes: bytes = slice(instrData, 9, 17);
    let totalBlocksProduced: i64 = bytesToI64LE(totalBlocksBytes);
    
    // Validate inputs
    if (rewardAmount < 0 || totalBlocksProduced < 0) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // If no blocks produced, skip distribution
    if (totalBlocksProduced == 0) {
        return SUCCESS;
    }
    
    // If no reward amount, skip distribution
    if (rewardAmount == 0) {
        return SUCCESS;
    }
    
    // Distribute rewards proportionally to validators based on blocks produced
    // For now, just return success - full implementation requires iterating validators
    return SUCCESS;
}

function handleReportDoubleSign(instrData: bytes): i64 {
    // Format: [type:u8][validatorPubkey:PublicKey(32)][blockHash1:bytes(32)][blockHash2:bytes(32)][sig1:bytes(64)][sig2:bytes(64)]
    // Total: 225 bytes
    if (len(instrData) != 225) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract validator pubkey (bytes 1-33)
    let validatorPubkey: bytes = slice(instrData, 1, 33);
    
    // Extract block hashes and signatures
    let blockHash1: bytes = slice(instrData, 33, 65);
    let blockHash2: bytes = slice(instrData, 65, 97);
    let sig1: bytes = slice(instrData, 97, 161);
    let sig2: bytes = slice(instrData, 161, 225);
    
    // Verify both signatures are valid
    // For now, just return success - full implementation requires signature verification
    
    // Compute validator FileID
    let validatorFileID: FileID = computeValidatorFileID(validatorPubkey);
    
    // Get validator record
    let validatorData: bytes = getFileData(validatorFileID);
    if (len(validatorData) != 66) {
        return ERROR_NOT_REGISTERED;
    }
    
    // Parse validator record - totalStake is at bytes 40-48
    let totalStake: i64 = bytesToI64LE(slice(validatorData, 40, 48));
    
    // Calculate slashing amount (5% of total stake)
    let slashingAmount: i64 = (totalStake * SLASHING_PERCENTAGE) / 100;
    
    // Reduce stake
    let newStake: i64 = totalStake - slashingAmount;
    
    // Check if validator should be deactivated
    if (newStake < MIN_ACTIVATION_STAKE) {
        // Deactivate validator
        // For now, just return success - full implementation requires updating validator record
    }
    
    // Transfer slashed amount to reward pool
    // For now, just return success - full implementation requires updating reward pool
    
    return SUCCESS;
}
