/*
 *  Copyright 2010 Barend Scholtus
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
package spaceinvaders;

/**
 *
 * @author Barend
 */
public class I8080OpInfo {
    public static final int OPND_NONE      = 0;
    public static final int OPND_R_R       = 1;
    public static final int OPND_M_R       = 2;
    public static final int OPND_R_M       = 3;
    public static final int OPND_R_DATA    = 4;
    public static final int OPND_M_DATA    = 5;
    public static final int OPND_RP_DATA16 = 6;
    public static final int OPND_RP        = 7;
    public static final int OPND_ADDR      = 8;
    public static final int OPND_R         = 9;

    private String mnemonic;
    private String description;
    private int cycles;
    private int operandType;

    private I8080OpInfo(String mnemonic, String description, int cyclesSmall) {
        this(mnemonic, description, cyclesSmall, I8080OpInfo.OPND_NONE);
    }

    private I8080OpInfo(String mnemonic, String description, int cyclesSmall, int operandType) {
        this.mnemonic = mnemonic;
        this.description = description;
        this.cycles = cyclesSmall;
        this.operandType = operandType;
    }

    public String getMnemonic() {
        return mnemonic;
    }

    public String getDescription() {
        return description;
    }

    public int getCycles() {
        return cycles;
    }

    public int getOperandType() {
        return operandType;
    }

    public static String getOperandsAsString(int[] memory, int opAddr) {
        int opcode = memory[opAddr];
        int data;
        StringBuilder sb = new StringBuilder();
        switch (opInfo[opcode].getOperandType()) {
            case OPND_R_R:
                return sb.append(intToRegister((opcode >> 3) & 0x07))
                         .append(' ').append(intToRegister(opcode & 0x07)).toString();
            case OPND_M_R:
                return sb.append('M')
                         .append(' ').append(intToRegister(opcode & 0x07)).toString();
            case OPND_R_M:
                return sb.append(intToRegister((opcode >> 3) & 0x07))
                         .append(' ').append('M').toString();
            case OPND_R_DATA:
                data = memory[(opAddr+1) % memory.length];
                return sb.append(intToRegister((opcode >> 3) & 0x07))
                         .append(' ').append(toHexString(data, 2)).toString();
            case OPND_M_DATA:
                data = memory[(opAddr+1) % memory.length];
                return sb.append('M').append(' ')
                         .append(toHexString(data, 2)).toString();
            case OPND_RP_DATA16:
                switch ((opcode >> 4) & 0x2) {
                    case 0: sb.append('B'); break;
                    case 1: sb.append('D'); break;
                    case 2: sb.append('H'); break;
                    case 3: sb.append('S').append('P'); break;
                }
                sb.append(' ');
                data = memory[(opAddr+1) % memory.length]
                            | (memory[(opAddr+2) % memory.length] << 8);
                sb.append(toHexString(data, 4));
                return sb.toString();
            case OPND_RP:
                switch ((opcode >> 4) & 0x2) {
                    case 0: sb.append('B'); break;
                    case 1: sb.append('D'); break;
                    case 2: sb.append('H'); break;
                    case 3: sb.append('S').append('P'); break;
                }
                return sb.toString();
            case OPND_ADDR:
                data = memory[(opAddr+1) % memory.length]
                            | (memory[(opAddr+2) % memory.length] << 8);
                return sb.append(toHexString(data, 4)).toString();
            case OPND_R:
                return sb.append(intToRegister((opcode >> 3) & 0x07)).toString();
        }
        return "";
    }

    public static char intToRegister(int reg) {
        switch (reg) {
            case 0: return 'B';
            case 1: return 'C';
            case 2: return 'D';
            case 3: return 'E';
            case 4: return 'H';
            case 5: return 'L';
            case 6: return 'M';
            case 7: return 'A';
        }
        return '?';
    }

    public static String toHexString(int i, int minLength) {
        StringBuilder sb = new StringBuilder();
        while (i != 0) {
            switch (i & 0xf) {
                case 0x0: sb.append('0'); break;
                case 0x1: sb.append('1'); break;
                case 0x2: sb.append('2'); break;
                case 0x3: sb.append('3'); break;
                case 0x4: sb.append('4'); break;
                case 0x5: sb.append('5'); break;
                case 0x6: sb.append('6'); break;
                case 0x7: sb.append('7'); break;
                case 0x8: sb.append('8'); break;
                case 0x9: sb.append('9'); break;
                case 0xa: sb.append('A'); break;
                case 0xb: sb.append('B'); break;
                case 0xc: sb.append('C'); break;
                case 0xd: sb.append('D'); break;
                case 0xe: sb.append('E'); break;
                case 0xf: sb.append('F'); break;
            }
            i >>>= 4;
        }
        while (sb.length() < minLength) {
            sb.append('0');
        }
        sb.reverse();
        return sb.toString();
    }

    // 0x00-0x0f
    private static final I8080OpInfo op0x00 = new I8080OpInfo("NOP", "No-operation", 4);
    private static final I8080OpInfo op0x01 = new I8080OpInfo("LXI", "Load immediate register Pair B & C", 10, I8080OpInfo.OPND_RP_DATA16);
    private static final I8080OpInfo op0x02 = new I8080OpInfo("STAX", "Store A indirect", 7, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x03 = new I8080OpInfo("INX", "Increment B & C registers", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x04 = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x05 = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x06 = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x07 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x08 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x09 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x0a = new I8080OpInfo("LDAX", "Load A indirect", 7, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x0b = new I8080OpInfo("DCX", "Decrement B & C", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x0c = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x0d = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x0e = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x0f = new I8080OpInfo("", "", 0, 0);
    // 0x10-0x1f
    private static final I8080OpInfo op0x10 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x11 = new I8080OpInfo("LXI", "Load immediate register Pair D & E", 10, I8080OpInfo.OPND_RP_DATA16);
    private static final I8080OpInfo op0x12 = new I8080OpInfo("STAX", "Store A indirect", 7, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x13 = new I8080OpInfo("INX", "Increment D & E registers", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x14 = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x15 = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x16 = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x17 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x18 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x19 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x1a = new I8080OpInfo("LDAX", "Load A indirect", 7, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x1b = new I8080OpInfo("DCX", "Decrement D & E", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x1c = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x1d = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x1e = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x1f = new I8080OpInfo("", "", 0, 0);
    // 0x20-0x2f
    private static final I8080OpInfo op0x20 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x21 = new I8080OpInfo("LXI", "Load immediate register Pair H & L", 10, I8080OpInfo.OPND_RP_DATA16);
    private static final I8080OpInfo op0x22 = new I8080OpInfo("SHLD", "Store H & L direct", 16, I8080OpInfo.OPND_ADDR);
    private static final I8080OpInfo op0x23 = new I8080OpInfo("INX", "Increment H & L registers", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x24 = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x25 = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x26 = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x27 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x28 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x29 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x2a = new I8080OpInfo("LHLD", "Load H & L direct", 16, I8080OpInfo.OPND_ADDR);
    private static final I8080OpInfo op0x2b = new I8080OpInfo("DCX", "Decrement H & L", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x2c = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x2d = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x2e = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x2f = new I8080OpInfo("", "", 0, 0);
    // 0x30-0x3f
    private static final I8080OpInfo op0x30 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x31 = new I8080OpInfo("LXI", "Load immediate stack pointer", 10, I8080OpInfo.OPND_RP_DATA16);
    private static final I8080OpInfo op0x32 = new I8080OpInfo("STA", "Store A direct", 13, I8080OpInfo.OPND_ADDR);
    private static final I8080OpInfo op0x33 = new I8080OpInfo("INX", "Increment stack pointer", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x34 = new I8080OpInfo("INR", "Increment memory", 10, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x35 = new I8080OpInfo("DCR", "Decrement memory", 10, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x36 = new I8080OpInfo("MVI", "Move immediate memory", 10, I8080OpInfo.OPND_M_DATA);
    private static final I8080OpInfo op0x37 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x38 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x39 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x3a = new I8080OpInfo("LDA", "Load A direct", 13, I8080OpInfo.OPND_ADDR);
    private static final I8080OpInfo op0x3b = new I8080OpInfo("DCX", "Decrement stack pointer", 5, I8080OpInfo.OPND_RP);
    private static final I8080OpInfo op0x3c = new I8080OpInfo("INR", "Increment register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x3d = new I8080OpInfo("DCR", "Decrement register", 5, I8080OpInfo.OPND_R);
    private static final I8080OpInfo op0x3e = new I8080OpInfo("MVI", "Move immediate register", 7, I8080OpInfo.OPND_R_DATA);
    private static final I8080OpInfo op0x3f = new I8080OpInfo("", "", 0, 0);
    // 0x40-0x4f
    private static final I8080OpInfo op0x40 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x41 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x42 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x43 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x44 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x45 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x46 = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x47 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x48 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x49 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x4a = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x4b = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x4c = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x4d = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x4e = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x4f = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    // 0x50-0x5f
    private static final I8080OpInfo op0x50 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x51 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x52 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x53 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x54 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x55 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x56 = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x57 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x58 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x59 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x5a = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x5b = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x5c = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x5d = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x5e = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x5f = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    // 0x60-0x6f
    private static final I8080OpInfo op0x60 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x61 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x62 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x63 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x64 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x65 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x66 = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x67 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x68 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x69 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x6a = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x6b = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x6c = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x6d = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x6e = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x6f = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    // 0x70-0x7f
    private static final I8080OpInfo op0x70 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x71 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x72 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x73 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x74 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x75 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x76 = new I8080OpInfo("HLT", "Halt", 5);
    private static final I8080OpInfo op0x77 = new I8080OpInfo("MOV", "Move register to memory", 5, I8080OpInfo.OPND_M_R);
    private static final I8080OpInfo op0x78 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x79 = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x7a = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x7b = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x7c = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x7d = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    private static final I8080OpInfo op0x7e = new I8080OpInfo("MOV", "Move memory to register", 5, I8080OpInfo.OPND_R_M);
    private static final I8080OpInfo op0x7f = new I8080OpInfo("MOV", "Move register to register", 5, I8080OpInfo.OPND_R_R);
    // 0x80-0x8f
    private static final I8080OpInfo op0x80 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x81 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x82 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x83 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x84 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x85 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x86 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x87 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x88 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x89 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x8a = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x8b = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x8c = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x8d = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x8e = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x8f = new I8080OpInfo("", "", 0, 0);
    // 0x90-0x9f
    private static final I8080OpInfo op0x90 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x91 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x92 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x93 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x94 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x95 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x96 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x97 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x98 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x99 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x9a = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x9b = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x9c = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x9d = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x9e = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0x9f = new I8080OpInfo("", "", 0, 0);
    // 0xa0-0xaf
    private static final I8080OpInfo op0xa0 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa1 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa2 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa3 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa4 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa5 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa6 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa7 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa8 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xa9 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xaa = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xab = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xac = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xad = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xae = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xaf = new I8080OpInfo("", "", 0, 0);
    // 0xb0-0xbf
    private static final I8080OpInfo op0xb0 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb1 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb2 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb3 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb4 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb5 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb6 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb7 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb8 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xb9 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xba = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xbb = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xbc = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xbd = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xbe = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xbf = new I8080OpInfo("", "", 0, 0);
    // 0xc0-0xcf
    private static final I8080OpInfo op0xc0 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc1 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc2 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc3 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc4 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc5 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc6 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc7 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc8 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xc9 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xca = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xcb = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xcc = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xcd = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xce = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xcf = new I8080OpInfo("", "", 0, 0);
    // 0xd0-0xdf
    private static final I8080OpInfo op0xd0 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd1 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd2 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd3 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd4 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd5 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd6 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd7 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd8 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xd9 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xda = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xdb = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xdc = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xdd = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xde = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xdf = new I8080OpInfo("", "", 0, 0);
    // 0xe0-0xef
    private static final I8080OpInfo op0xe0 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe1 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe2 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe3 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe4 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe5 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe6 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe7 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe8 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xe9 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xea = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xeb = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xec = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xed = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xee = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xef = new I8080OpInfo("", "", 0, 0);
    // 0xf0-0xff
    private static final I8080OpInfo op0xf0 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf1 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf2 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf3 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf4 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf5 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf6 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf7 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf8 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xf9 = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xfa = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xfb = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xfc = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xfd = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xfe = new I8080OpInfo("", "", 0, 0);
    private static final I8080OpInfo op0xff = new I8080OpInfo("", "", 0, 0);

    public static final I8080OpInfo[] opInfo = new I8080OpInfo[] {
        op0x00, op0x01, op0x02, op0x03, op0x04, op0x05, op0x06, op0x07, op0x08, op0x09, op0x0a, op0x0b, op0x0c, op0x0d, op0x0e, op0x0f,
        op0x10, op0x11, op0x12, op0x13, op0x14, op0x15, op0x16, op0x17, op0x18, op0x19, op0x1a, op0x1b, op0x1c, op0x1d, op0x1e, op0x1f,
        op0x20, op0x21, op0x22, op0x23, op0x24, op0x25, op0x26, op0x27, op0x28, op0x29, op0x2a, op0x2b, op0x2c, op0x2d, op0x2e, op0x2f,
        op0x30, op0x31, op0x32, op0x33, op0x34, op0x35, op0x36, op0x37, op0x38, op0x39, op0x3a, op0x3b, op0x3c, op0x3d, op0x3e, op0x3f,
        op0x40, op0x41, op0x42, op0x43, op0x44, op0x45, op0x46, op0x47, op0x48, op0x49, op0x4a, op0x4b, op0x4c, op0x4d, op0x4e, op0x4f,
        op0x50, op0x51, op0x52, op0x53, op0x54, op0x55, op0x56, op0x57, op0x58, op0x59, op0x5a, op0x5b, op0x5c, op0x5d, op0x5e, op0x5f,
        op0x60, op0x61, op0x62, op0x63, op0x64, op0x65, op0x66, op0x67, op0x68, op0x69, op0x6a, op0x6b, op0x6c, op0x6d, op0x6e, op0x6f,
        op0x70, op0x71, op0x72, op0x73, op0x74, op0x75, op0x76, op0x77, op0x78, op0x79, op0x7a, op0x7b, op0x7c, op0x7d, op0x7e, op0x7f,
        op0x80, op0x81, op0x82, op0x83, op0x84, op0x85, op0x86, op0x87, op0x88, op0x89, op0x8a, op0x8b, op0x8c, op0x8d, op0x8e, op0x8f,
        op0x90, op0x91, op0x92, op0x93, op0x94, op0x95, op0x96, op0x97, op0x98, op0x99, op0x9a, op0x9b, op0x9c, op0x9d, op0x9e, op0x9f,
        op0xa0, op0xa1, op0xa2, op0xa3, op0xa4, op0xa5, op0xa6, op0xa7, op0xa8, op0xa9, op0xaa, op0xab, op0xac, op0xad, op0xae, op0xaf,
        op0xb0, op0xb1, op0xb2, op0xb3, op0xb4, op0xb5, op0xb6, op0xb7, op0xb8, op0xb9, op0xba, op0xbb, op0xbc, op0xbd, op0xbe, op0xbf,
        op0xc0, op0xc1, op0xc2, op0xc3, op0xc4, op0xc5, op0xc6, op0xc7, op0xc8, op0xc9, op0xca, op0xcb, op0xcc, op0xcd, op0xce, op0xcf,
        op0xd0, op0xd1, op0xd2, op0xd3, op0xd4, op0xd5, op0xd6, op0xd7, op0xd8, op0xd9, op0xda, op0xdb, op0xdc, op0xdd, op0xde, op0xdf,
        op0xe0, op0xe1, op0xe2, op0xe3, op0xe4, op0xe5, op0xe6, op0xe7, op0xe8, op0xe9, op0xea, op0xeb, op0xec, op0xed, op0xee, op0xef,
        op0xf0, op0xf1, op0xf2, op0xf3, op0xf4, op0xf5, op0xf6, op0xf7, op0xf8, op0xf9, op0xfa, op0xfb, op0xfc, op0xfd, op0xfe, op0xff
    };
}
