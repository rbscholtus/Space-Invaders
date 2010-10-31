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
package i8080;

import java.util.Formatter;

/**
 *
 * @author Barend Scholtus
 */
public class I8080OpInfo {
//    public static final int OPND_NONE      = 0;
//    public static final int OPND_R_R       = 1;
//    public static final int OPND_M_R       = 2;
//    public static final int OPND_R_M       = 3;
//    public static final int OPND_R_DATA    = 4;
//    public static final int OPND_M_DATA    = 5;
//    public static final int OPND_RP_DATA16 = 6;
//    public static final int OPND_RP        = 7;
//    public static final int OPND_ADDR      = 8;
//    public static final int OPND_R         = 9;

    public final String mnemonic;
    public final String operands;
    public final int length;
    public final int cyclesShort;
    public final int cyclesLong;
    public final boolean affectS;
    public final boolean affectZ;
    public final boolean affectA;
    public final boolean affectP;
    public final boolean affectC;
    public final String description;
    public final int operandType;

    /**
     *
     * @param mnemonic
     * @param operands
     * @param length
     * @param cyclesShort
     * @param cyclesLong
     * @param S
     * @param Z
     * @param A
     * @param P
     * @param C
     * @param description
     * @param operandType
     */
    private I8080OpInfo(String mnemonic, String operands,
            int length, int cyclesShort, int cyclesLong,
            boolean S, boolean Z, boolean A, boolean P, boolean C,
            String description, int operandType) {
        this.mnemonic = mnemonic;
        this.operands = operands;
        this.length = length;
        this.cyclesShort = cyclesShort;
        this.cyclesLong = cyclesLong;
        this.affectS = S;
        this.affectZ = Z;
        this.affectA = A;
        this.affectP = P;
        this.affectC = C;
        this.description = description;
        this.operandType = operandType;
    }

//    public static String getOperandsAsString(int[] memory, int opAddr) {
//        int opcode = memory[opAddr];
//        int data;
//        StringBuilder sb = new StringBuilder();
//        switch (opInfo[opcode].getOperandType()) {
//            case OPND_R_R:
//                return sb.append(intToRegister((opcode >> 3) & 0x07))
//                         .append(' ').append(intToRegister(opcode & 0x07)).toString();
//            case OPND_M_R:
//                return sb.append('M')
//                         .append(' ').append(intToRegister(opcode & 0x07)).toString();
//            case OPND_R_M:
//                return sb.append(intToRegister((opcode >> 3) & 0x07))
//                         .append(' ').append('M').toString();
//            case OPND_R_DATA:
//                data = memory[(opAddr+1) % memory.length];
//                return sb.append(intToRegister((opcode >> 3) & 0x07))
//                         .append(' ').append(toHexString(data, 2)).toString();
//            case OPND_M_DATA:
//                data = memory[(opAddr+1) % memory.length];
//                return sb.append('M').append(' ')
//                         .append(toHexString(data, 2)).toString();
//            case OPND_RP_DATA16:
//                switch ((opcode >> 4) & 0x2) {
//                    case 0: sb.append('B'); break;
//                    case 1: sb.append('D'); break;
//                    case 2: sb.append('H'); break;
//                    case 3: sb.append('S').append('P'); break;
//                }
//                sb.append(' ');
//                data = memory[(opAddr+1) % memory.length]
//                            | (memory[(opAddr+2) % memory.length] << 8);
//                sb.append(toHexString(data, 4));
//                return sb.toString();
//            case OPND_RP:
//                switch ((opcode >> 4) & 0x2) {
//                    case 0: sb.append('B'); break;
//                    case 1: sb.append('D'); break;
//                    case 2: sb.append('H'); break;
//                    case 3: sb.append('S').append('P'); break;
//                }
//                return sb.toString();
//            case OPND_ADDR:
//                data = memory[(opAddr+1) % memory.length]
//                            | (memory[(opAddr+2) % memory.length] << 8);
//                return sb.append(toHexString(data, 4)).toString();
//            case OPND_R:
//                return sb.append(intToRegister((opcode >> 3) & 0x07)).toString();
//        }
//        return "";
//    }

    /**
     *
     * @param reg
     * @return
     */
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
        throw new IllegalArgumentException("Not a valid register: " + reg);
    }

    /**
     * 
     * @param cpu
     * @param ctx
     * @return
     */
    public static String debugString(I8080 cpu, I8080Context ctx) {
        StringBuilder sb = new StringBuilder(80);
        Formatter f = new Formatter(sb);

        I8080OpInfo op = opCode[ctx.read(cpu.PC)];

        f.format("PC=%04x SP=%04x A=%02x B=%02x C=%02x D=%02x E=%02x H=%02x L=%02x %d%d%d%d%d op:%5s %s",
                cpu.PC, cpu.SP,
                cpu.A, cpu.B, cpu.C, cpu.D, cpu.E, cpu.H, cpu.L,
                cpu.Carry ? 1 : 0, cpu.Zero ? 1 : 0, cpu.Sign ? 1 : 0, cpu.Parity ? 1 : 0, cpu.AuxCarry ? 1 : 0,
                op.mnemonic, op.operands);

        return sb.toString();
    }

    // 0x00-0x0f
    private static final I8080OpInfo op0x00 = new I8080OpInfo("NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x01 = new I8080OpInfo("LXI", "B,d16", 3, 10, 10, false, false, false, false, false, "Load immediate BC", 0);
    private static final I8080OpInfo op0x02 = new I8080OpInfo("STAX", "B", 1, 7, 7, false, false, false, false, false, "Store A indirect", 0);
    private static final I8080OpInfo op0x03 = new I8080OpInfo("INX", "B", 1, 5, 5, false, false, false, false, false, "Increment BC", 0);
    private static final I8080OpInfo op0x04 = new I8080OpInfo("INR", "B", 1, 5, 5, true, true, true, true, false, "Increment register B", 0);
    private static final I8080OpInfo op0x05 = new I8080OpInfo("DCR", "B", 1, 5, 5, true, true, true, true, false, "Decrement register B", 0);
    private static final I8080OpInfo op0x06 = new I8080OpInfo("MVI", "B,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to B", 0);
    private static final I8080OpInfo op0x07 = new I8080OpInfo("RLC", "", 1, 4, 4, false, false, false, false, true, "Rotate A left", 0);
    private static final I8080OpInfo op0x08 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x09 = new I8080OpInfo("DAD", "B", 1, 10, 10, false, false, false, false, true, "Add BC to HL", 0);
    private static final I8080OpInfo op0x0a = new I8080OpInfo("LDAX", "B", 1, 7, 7, false, false, false, false, false, "Load A indirect via BC", 0);
    private static final I8080OpInfo op0x0b = new I8080OpInfo("DCX", "B", 1, 5, 5, false, false, false, false, false, "Decrement BC", 0);
    private static final I8080OpInfo op0x0c = new I8080OpInfo("INR", "C", 1, 5, 5, true, true, true, true, false, "Increment register C", 0);
    private static final I8080OpInfo op0x0d = new I8080OpInfo("DCR", "C", 1, 5, 5, true, true, true, true, false, "Decrement register C", 0);
    private static final I8080OpInfo op0x0e = new I8080OpInfo("MVI", "C,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to C", 0);
    private static final I8080OpInfo op0x0f = new I8080OpInfo("RRC", "", 1, 4, 4, false, false, false, false, true, "Rotate A right", 0);
    // 0x10-0x1f
    private static final I8080OpInfo op0x10 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x11 = new I8080OpInfo("LXI", "D,d16", 3, 10, 10, false, false, false, false, false, "Load immediate DE", 0);
    private static final I8080OpInfo op0x12 = new I8080OpInfo("STAX", "D", 1, 7, 7, false, false, false, false, false, "Store A indirect", 0);
    private static final I8080OpInfo op0x13 = new I8080OpInfo("INX", "D", 1, 5, 5, false, false, false, false, false, "Increment DE", 0);
    private static final I8080OpInfo op0x14 = new I8080OpInfo("INR", "D", 1, 5, 5, true, true, true, true, false, "Increment register D", 0);
    private static final I8080OpInfo op0x15 = new I8080OpInfo("DCR", "D", 1, 5, 5, true, true, true, true, false, "Decrement register D", 0);
    private static final I8080OpInfo op0x16 = new I8080OpInfo("MVI", "D,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to D", 0);
    private static final I8080OpInfo op0x17 = new I8080OpInfo("RAL", "", 1, 4, 4, false, false, false, false, true, "Rotate A left through carry", 0);
    private static final I8080OpInfo op0x18 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x19 = new I8080OpInfo("DAD", "D", 1, 10, 10, false, false, false, false, true, "Add DE to HL", 0);
    private static final I8080OpInfo op0x1a = new I8080OpInfo("LDAX", "D", 1, 7, 7, false, false, false, false, false, "Load A indirect via DE", 0);
    private static final I8080OpInfo op0x1b = new I8080OpInfo("DCX", "D", 1, 5, 5, false, false, false, false, false, "Decrement DE", 0);
    private static final I8080OpInfo op0x1c = new I8080OpInfo("INR", "E", 1, 5, 5, true, true, true, true, false, "Increment register E", 0);
    private static final I8080OpInfo op0x1d = new I8080OpInfo("DCR", "E", 1, 5, 5, true, true, true, true, false, "Decrement register E", 0);
    private static final I8080OpInfo op0x1e = new I8080OpInfo("MVI", "E,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to E", 0);
    private static final I8080OpInfo op0x1f = new I8080OpInfo("RAR", "", 1, 4, 4, false, false, false, false, true, "Rotate A right through carry", 0);
    // 0x20-0x2f
    private static final I8080OpInfo op0x20 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x21 = new I8080OpInfo("LXI", "H,d16", 3, 10, 10, false, false, false, false, false, "Load immediate HL", 0);
    private static final I8080OpInfo op0x22 = new I8080OpInfo("SHLD", "a16", 3, 16, 16, false, false, false, false, false, "Store HL direct", 0);
    private static final I8080OpInfo op0x23 = new I8080OpInfo("INX", "H", 1, 5, 5, false, false, false, false, false, "Increment HL", 0);
    private static final I8080OpInfo op0x24 = new I8080OpInfo("INR", "H", 1, 5, 5, true, true, true, true, false, "Increment register H", 0);
    private static final I8080OpInfo op0x25 = new I8080OpInfo("DCR", "H", 1, 5, 5, true, true, true, true, false, "Decrement register H", 0);
    private static final I8080OpInfo op0x26 = new I8080OpInfo("MVI", "H,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to H", 0);
    private static final I8080OpInfo op0x27 = new I8080OpInfo("DAA", "", 1, 4, 4, true, true, true, true, true, "Decimal adjust A", 0);
    private static final I8080OpInfo op0x28 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x29 = new I8080OpInfo("DAD", "H", 1, 10, 10, false, false, false, false, true, "Add HL to HL", 0);
    private static final I8080OpInfo op0x2a = new I8080OpInfo("LHLD", "a16", 3, 16, 16, false, false, false, false, false, "Load HL direct", 0);
    private static final I8080OpInfo op0x2b = new I8080OpInfo("DCX", "H", 1, 5, 5, false, false, false, false, false, "Decrement HL", 0);
    private static final I8080OpInfo op0x2c = new I8080OpInfo("INR", "L", 1, 5, 5, true, true, true, true, false, "Increment register L", 0);
    private static final I8080OpInfo op0x2d = new I8080OpInfo("DCR", "L", 1, 5, 5, true, true, true, true, false, "Decrement register L", 0);
    private static final I8080OpInfo op0x2e = new I8080OpInfo("MVI", "L,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to L", 0);
    private static final I8080OpInfo op0x2f = new I8080OpInfo("CMA", "", 1, 4, 4, false, false, false, false, false, "Complement A", 0);
    // 0x30-0x3f
    private static final I8080OpInfo op0x30 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x31 = new I8080OpInfo("LXI", "SP,d16", 3, 10, 10, false, false, false, false, false, "Load immediate stack pointer", 0);
    private static final I8080OpInfo op0x32 = new I8080OpInfo("STA", "a16", 3, 13, 13, false, false, false, false, false, "Store accumulator direct", 0);
    private static final I8080OpInfo op0x33 = new I8080OpInfo("INX", "SP", 1, 5, 5, false, false, false, false, false, "Increment stack pointer", 0);
    private static final I8080OpInfo op0x34 = new I8080OpInfo("INR", "M", 1, 10, 10, true, true, true, true, false, "Increment memory", 0);
    private static final I8080OpInfo op0x35 = new I8080OpInfo("DCR", "M", 1, 10, 10, true, true, true, true, false, "Decrement memory", 0);
    private static final I8080OpInfo op0x36 = new I8080OpInfo("MVI", "M,d8", 2, 10, 10, false, false, false, false, false, "Move immediate to memory", 0);
    private static final I8080OpInfo op0x37 = new I8080OpInfo("STC", "", 1, 4, 4, false, false, false, false, true, "Set carry", 0);
    private static final I8080OpInfo op0x38 = new I8080OpInfo("*NOP", "", 1, 4, 4, false, false, false, false, false, "No-operation", 0);
    private static final I8080OpInfo op0x39 = new I8080OpInfo("DAD", "SP", 1, 10, 10, false, false, false, false, true, "Add SP to HL", 0);
    private static final I8080OpInfo op0x3a = new I8080OpInfo("LDA", "a16", 3, 13, 13, false, false, false, false, false, "Load A direct", 0);
    private static final I8080OpInfo op0x3b = new I8080OpInfo("DCX", "SP", 1, 5, 5, false, false, false, false, false, "Decrement stack pointer", 0);
    private static final I8080OpInfo op0x3c = new I8080OpInfo("INR", "A", 1, 5, 5, true, true, true, true, false, "Increment register A", 0);
    private static final I8080OpInfo op0x3d = new I8080OpInfo("DCR", "A", 1, 5, 5, true, true, true, true, false, "Decrement register A", 0);
    private static final I8080OpInfo op0x3e = new I8080OpInfo("MVI", "A,d8", 2, 7, 7, false, false, false, false, false, "Move immediate to A", 0);
    private static final I8080OpInfo op0x3f = new I8080OpInfo("CMC", "", 1, 4, 4, false, false, false, false, true, "Complement carry", 0);
    // 0x40-0x4f
    private static final I8080OpInfo op0x40 = new I8080OpInfo("MOV", "B,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x41 = new I8080OpInfo("MOV", "B,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x42 = new I8080OpInfo("MOV", "B,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x43 = new I8080OpInfo("MOV", "B,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x44 = new I8080OpInfo("MOV", "B,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x45 = new I8080OpInfo("MOV", "B,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x46 = new I8080OpInfo("MOV", "B,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x47 = new I8080OpInfo("MOV", "B,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x48 = new I8080OpInfo("MOV", "C,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x49 = new I8080OpInfo("MOV", "C,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x4a = new I8080OpInfo("MOV", "C,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x4b = new I8080OpInfo("MOV", "C,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x4c = new I8080OpInfo("MOV", "C,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x4d = new I8080OpInfo("MOV", "C,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x4e = new I8080OpInfo("MOV", "C,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x4f = new I8080OpInfo("MOV", "C,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    // 0x50-0x5f
    private static final I8080OpInfo op0x50 = new I8080OpInfo("MOV", "D,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x51 = new I8080OpInfo("MOV", "D,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x52 = new I8080OpInfo("MOV", "D,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x53 = new I8080OpInfo("MOV", "D,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x54 = new I8080OpInfo("MOV", "D,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x55 = new I8080OpInfo("MOV", "D,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x56 = new I8080OpInfo("MOV", "D,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x57 = new I8080OpInfo("MOV", "D,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x58 = new I8080OpInfo("MOV", "E,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x59 = new I8080OpInfo("MOV", "E,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x5a = new I8080OpInfo("MOV", "E,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x5b = new I8080OpInfo("MOV", "E,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x5c = new I8080OpInfo("MOV", "E,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x5d = new I8080OpInfo("MOV", "E,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x5e = new I8080OpInfo("MOV", "E,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x5f = new I8080OpInfo("MOV", "E,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    // 0x60-0x6f
    private static final I8080OpInfo op0x60 = new I8080OpInfo("MOV", "H,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x61 = new I8080OpInfo("MOV", "H,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x62 = new I8080OpInfo("MOV", "H,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x63 = new I8080OpInfo("MOV", "H,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x64 = new I8080OpInfo("MOV", "H,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x65 = new I8080OpInfo("MOV", "H,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x66 = new I8080OpInfo("MOV", "H,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x67 = new I8080OpInfo("MOV", "H,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x68 = new I8080OpInfo("MOV", "L,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x69 = new I8080OpInfo("MOV", "L,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x6a = new I8080OpInfo("MOV", "L,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x6b = new I8080OpInfo("MOV", "L,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x6c = new I8080OpInfo("MOV", "L,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x6d = new I8080OpInfo("MOV", "L,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x6e = new I8080OpInfo("MOV", "L,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x6f = new I8080OpInfo("MOV", "L,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    // 0x70-0x7f
    private static final I8080OpInfo op0x70 = new I8080OpInfo("MOV", "M,B", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x71 = new I8080OpInfo("MOV", "M,C", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x72 = new I8080OpInfo("MOV", "M,D", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x73 = new I8080OpInfo("MOV", "M,E", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x74 = new I8080OpInfo("MOV", "M,H", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x75 = new I8080OpInfo("MOV", "M,L", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x76 = new I8080OpInfo("HLT", "", 1, 7, 7, false, false, false, false, false, "Halt", 0);
    private static final I8080OpInfo op0x77 = new I8080OpInfo("MOV", "M,A", 1, 7, 7, false, false, false, false, false, "Move register to memory", 0);
    private static final I8080OpInfo op0x78 = new I8080OpInfo("MOV", "A,B", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x79 = new I8080OpInfo("MOV", "A,C", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x7a = new I8080OpInfo("MOV", "A,D", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x7b = new I8080OpInfo("MOV", "A,E", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x7c = new I8080OpInfo("MOV", "A,H", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x7d = new I8080OpInfo("MOV", "A,L", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    private static final I8080OpInfo op0x7e = new I8080OpInfo("MOV", "A,M", 1, 7, 7, false, false, false, false, false, "Move memory to register", 0);
    private static final I8080OpInfo op0x7f = new I8080OpInfo("MOV", "A,A", 1, 5, 5, false, false, false, false, false, "Move register to register", 0);
    // 0x80-0x8f
    private static final I8080OpInfo op0x80 = new I8080OpInfo("ADD", "B", 1, 4, 4, true, true, true, true, true, "Add B to A", 0);
    private static final I8080OpInfo op0x81 = new I8080OpInfo("ADD", "C", 1, 4, 4, true, true, true, true, true, "Add C to A", 0);
    private static final I8080OpInfo op0x82 = new I8080OpInfo("ADD", "D", 1, 4, 4, true, true, true, true, true, "Add D to A", 0);
    private static final I8080OpInfo op0x83 = new I8080OpInfo("ADD", "E", 1, 4, 4, true, true, true, true, true, "Add E to A", 0);
    private static final I8080OpInfo op0x84 = new I8080OpInfo("ADD", "H", 1, 4, 4, true, true, true, true, true, "Add H to A", 0);
    private static final I8080OpInfo op0x85 = new I8080OpInfo("ADD", "L", 1, 4, 4, true, true, true, true, true, "Add L to A", 0);
    private static final I8080OpInfo op0x86 = new I8080OpInfo("ADD", "M", 1, 7, 7, true, true, true, true, true, "Add memory to A", 0);
    private static final I8080OpInfo op0x87 = new I8080OpInfo("ADD", "A", 1, 4, 4, true, true, true, true, true, "Add A to A", 0);
    private static final I8080OpInfo op0x88 = new I8080OpInfo("ADC", "B", 1, 4, 4, true, true, true, true, true, "Add B to A with carry", 0);
    private static final I8080OpInfo op0x89 = new I8080OpInfo("ADC", "C", 1, 4, 4, true, true, true, true, true, "Add C to A with carry", 0);
    private static final I8080OpInfo op0x8a = new I8080OpInfo("ADC", "D", 1, 4, 4, true, true, true, true, true, "Add D to A with carry", 0);
    private static final I8080OpInfo op0x8b = new I8080OpInfo("ADC", "E", 1, 4, 4, true, true, true, true, true, "Add E to A with carry", 0);
    private static final I8080OpInfo op0x8c = new I8080OpInfo("ADC", "H", 1, 4, 4, true, true, true, true, true, "Add H to A with carry", 0);
    private static final I8080OpInfo op0x8d = new I8080OpInfo("ADC", "L", 1, 4, 4, true, true, true, true, true, "Add L to A with carry", 0);
    private static final I8080OpInfo op0x8e = new I8080OpInfo("ADC", "M", 1, 7, 7, true, true, true, true, true, "Add memory to A with carry", 0);
    private static final I8080OpInfo op0x8f = new I8080OpInfo("ADC", "A", 1, 4, 4, true, true, true, true, true, "Add A to A with carry", 0);
    // 0x90-0x9f
    private static final I8080OpInfo op0x90 = new I8080OpInfo("SUB", "B", 1, 4, 4, true, true, true, true, true, "Subtract B from A", 0);
    private static final I8080OpInfo op0x91 = new I8080OpInfo("SUB", "C", 1, 4, 4, true, true, true, true, true, "Subtract C from A", 0);
    private static final I8080OpInfo op0x92 = new I8080OpInfo("SUB", "D", 1, 4, 4, true, true, true, true, true, "Subtract D from A", 0);
    private static final I8080OpInfo op0x93 = new I8080OpInfo("SUB", "E", 1, 4, 4, true, true, true, true, true, "Subtract E from A", 0);
    private static final I8080OpInfo op0x94 = new I8080OpInfo("SUB", "H", 1, 4, 4, true, true, true, true, true, "Subtract H from A", 0);
    private static final I8080OpInfo op0x95 = new I8080OpInfo("SUB", "L", 1, 4, 4, true, true, true, true, true, "Subtract L from A", 0);
    private static final I8080OpInfo op0x96 = new I8080OpInfo("SUB", "M", 1, 7, 7, true, true, true, true, true, "Subtract memory from A", 0);
    private static final I8080OpInfo op0x97 = new I8080OpInfo("SUB", "A", 1, 4, 4, true, true, true, true, true, "Subtract A from A", 0);
    private static final I8080OpInfo op0x98 = new I8080OpInfo("SBB", "B", 1, 4, 4, true, true, true, true, true, "Subtract B from A with borrow", 0);
    private static final I8080OpInfo op0x99 = new I8080OpInfo("SBB", "C", 1, 4, 4, true, true, true, true, true, "Subtract C from A with borrow", 0);
    private static final I8080OpInfo op0x9a = new I8080OpInfo("SBB", "D", 1, 4, 4, true, true, true, true, true, "Subtract D from A with borrow", 0);
    private static final I8080OpInfo op0x9b = new I8080OpInfo("SBB", "E", 1, 4, 4, true, true, true, true, true, "Subtract E from A with borrow", 0);
    private static final I8080OpInfo op0x9c = new I8080OpInfo("SBB", "H", 1, 4, 4, true, true, true, true, true, "Subtract H from A with borrow", 0);
    private static final I8080OpInfo op0x9d = new I8080OpInfo("SBB", "L", 1, 4, 4, true, true, true, true, true, "Subtract L from A with borrow", 0);
    private static final I8080OpInfo op0x9e = new I8080OpInfo("SBB", "M", 1, 7, 7, true, true, true, true, true, "Subtract memory from A with borrow", 0);
    private static final I8080OpInfo op0x9f = new I8080OpInfo("SBB", "A", 1, 4, 4, true, true, true, true, true, "Subtract A from A with borrow", 0);
    // 0xa0-0xaf
    private static final I8080OpInfo op0xa0 = new I8080OpInfo("ANA", "B", 1, 4, 4, true, true, false, true, true, "AND B with A", 0);
    private static final I8080OpInfo op0xa1 = new I8080OpInfo("ANA", "C", 1, 4, 4, true, true, false, true, true, "AND C with A", 0);
    private static final I8080OpInfo op0xa2 = new I8080OpInfo("ANA", "D", 1, 4, 4, true, true, false, true, true, "AND D with A", 0);
    private static final I8080OpInfo op0xa3 = new I8080OpInfo("ANA", "E", 1, 4, 4, true, true, false, true, true, "AND E with A", 0);
    private static final I8080OpInfo op0xa4 = new I8080OpInfo("ANA", "H", 1, 4, 4, true, true, false, true, true, "AND H with A", 0);
    private static final I8080OpInfo op0xa5 = new I8080OpInfo("ANA", "L", 1, 4, 4, true, true, false, true, true, "AND L with A", 0);
    private static final I8080OpInfo op0xa6 = new I8080OpInfo("ANA", "M", 1, 7, 7, true, true, false, true, true, "AND memory with A", 0);
    private static final I8080OpInfo op0xa7 = new I8080OpInfo("ANA", "A", 1, 4, 4, true, true, false, true, true, "AND A with A", 0);
    private static final I8080OpInfo op0xa8 = new I8080OpInfo("XRA", "B", 1, 4, 4, true, true, false, true, true, "Exclusive OR B with A", 0);
    private static final I8080OpInfo op0xa9 = new I8080OpInfo("XRA", "C", 1, 4, 4, true, true, false, true, true, "Exclusive OR C with A", 0);
    private static final I8080OpInfo op0xaa = new I8080OpInfo("XRA", "D", 1, 4, 4, true, true, false, true, true, "Exclusive OR D with A", 0);
    private static final I8080OpInfo op0xab = new I8080OpInfo("XRA", "E", 1, 4, 4, true, true, false, true, true, "Exclusive OR E with A", 0);
    private static final I8080OpInfo op0xac = new I8080OpInfo("XRA", "H", 1, 4, 4, true, true, false, true, true, "Exclusive OR H with A", 0);
    private static final I8080OpInfo op0xad = new I8080OpInfo("XRA", "L", 1, 4, 4, true, true, false, true, true, "Exclusive OR L with A", 0);
    private static final I8080OpInfo op0xae = new I8080OpInfo("XRA", "M", 1, 7, 7, true, true, false, true, true, "Exclusive OR memory with A", 0);
    private static final I8080OpInfo op0xaf = new I8080OpInfo("XRA", "A", 1, 4, 4, true, true, false, true, true, "Exclusive OR A with A", 0);
    // 0xb0-0xbf
    private static final I8080OpInfo op0xb0 = new I8080OpInfo("ORA", "B", 1, 4, 4, true, true, false, true, true, "OR B with A", 0);
    private static final I8080OpInfo op0xb1 = new I8080OpInfo("ORA", "C", 1, 4, 4, true, true, false, true, true, "OR C with A", 0);
    private static final I8080OpInfo op0xb2 = new I8080OpInfo("ORA", "D", 1, 4, 4, true, true, false, true, true, "OR D with A", 0);
    private static final I8080OpInfo op0xb3 = new I8080OpInfo("ORA", "E", 1, 4, 4, true, true, false, true, true, "OR E with A", 0);
    private static final I8080OpInfo op0xb4 = new I8080OpInfo("ORA", "H", 1, 4, 4, true, true, false, true, true, "OR H with A", 0);
    private static final I8080OpInfo op0xb5 = new I8080OpInfo("ORA", "L", 1, 4, 4, true, true, false, true, true, "OR L with A", 0);
    private static final I8080OpInfo op0xb6 = new I8080OpInfo("ORA", "M", 1, 7, 7, true, true, false, true, true, "OR memory with A", 0);
    private static final I8080OpInfo op0xb7 = new I8080OpInfo("ORA", "A", 1, 4, 4, true, true, false, true, true, "OR A with A", 0);
    private static final I8080OpInfo op0xb8 = new I8080OpInfo("CMP", "B", 1, 4, 4, true, true, true, true, true, "Compare B with A", 0);
    private static final I8080OpInfo op0xb9 = new I8080OpInfo("CMP", "C", 1, 4, 4, true, true, true, true, true, "Compare C with A", 0);
    private static final I8080OpInfo op0xba = new I8080OpInfo("CMP", "D", 1, 4, 4, true, true, true, true, true, "Compare D with A", 0);
    private static final I8080OpInfo op0xbb = new I8080OpInfo("CMP", "E", 1, 4, 4, true, true, true, true, true, "Compare E with A", 0);
    private static final I8080OpInfo op0xbc = new I8080OpInfo("CMP", "H", 1, 4, 4, true, true, true, true, true, "Compare H with A", 0);
    private static final I8080OpInfo op0xbd = new I8080OpInfo("CMP", "L", 1, 4, 4, true, true, true, true, true, "Compare L with A", 0);
    private static final I8080OpInfo op0xbe = new I8080OpInfo("CMP", "M", 1, 7, 7, true, true, true, true, true, "Compare memory with A", 0);
    private static final I8080OpInfo op0xbf = new I8080OpInfo("CMP", "A", 1, 4, 4, true, true, true, true, true, "Compare A with A", 0);
    // 0xc0-0xcf
    private static final I8080OpInfo op0xc0 = new I8080OpInfo("RNZ", "", 1, 5, 11, false, false, false, false, false, "Return on non zero", 0);
    private static final I8080OpInfo op0xc1 = new I8080OpInfo("POP", "B", 1, 10, 10, false, false, false, false, false, "Pop BC off stack", 0);
    private static final I8080OpInfo op0xc2 = new I8080OpInfo("JNZ", "a16", 3, 10, 10, false, false, false, false, false, "Jump on no zero", 0);
    private static final I8080OpInfo op0xc3 = new I8080OpInfo("JMP", "a16", 3, 10, 10, false, false, false, false, false, "Jump unconditional", 0);
    private static final I8080OpInfo op0xc4 = new I8080OpInfo("CNZ", "a16", 3, 11, 17, false, false, false, false, false, "Call on no zero", 0);
    private static final I8080OpInfo op0xc5 = new I8080OpInfo("PUSH", "B", 1, 11, 11, false, false, false, false, false, "Push BC on stack", 0);
    private static final I8080OpInfo op0xc6 = new I8080OpInfo("ADI", "d8", 2, 7, 7, true, true, true, true, true, "Add immediate to A", 0);
    private static final I8080OpInfo op0xc7 = new I8080OpInfo("RST", "0", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    private static final I8080OpInfo op0xc8 = new I8080OpInfo("RZ", "", 1, 5, 11, false, false, false, false, false, "Return on zero", 0);
    private static final I8080OpInfo op0xc9 = new I8080OpInfo("RET", "", 1, 10, 10, false, false, false, false, false, "Return", 0);
    private static final I8080OpInfo op0xca = new I8080OpInfo("JZ", "a16", 3, 10, 10, false, false, false, false, false, "Jump on zero", 0);
    private static final I8080OpInfo op0xcb = new I8080OpInfo("*JMP", "a16", 3, 10, 10, false, false, false, false, false, "Jump unconditional", 0);
    private static final I8080OpInfo op0xcc = new I8080OpInfo("CZ", "a16", 3, 11, 17, false, false, false, false, false, "Call on zero", 0);
    private static final I8080OpInfo op0xcd = new I8080OpInfo("CALL", "a16", 3, 17, 17, false, false, false, false, false, "Call unconditional", 0);
    private static final I8080OpInfo op0xce = new I8080OpInfo("ACI", "d8", 2, 7, 7, true, true, true, true, true, "Add immediate to A with carry", 0);
    private static final I8080OpInfo op0xcf = new I8080OpInfo("RST", "1", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    // 0xd0-0xdf
    private static final I8080OpInfo op0xd0 = new I8080OpInfo("RNC", "", 1, 5, 11, false, false, false, false, false, "Return on no carry", 0);
    private static final I8080OpInfo op0xd1 = new I8080OpInfo("POP", "D", 1, 10, 10, false, false, false, false, false, "Pop DE off stack", 0);
    private static final I8080OpInfo op0xd2 = new I8080OpInfo("JNC", "a16", 3, 10, 10, false, false, false, false, false, "Jump on no carry", 0);
    private static final I8080OpInfo op0xd3 = new I8080OpInfo("OUT", "d8", 2, 10, 10, false, false, false, false, false, "", 0);
    private static final I8080OpInfo op0xd4 = new I8080OpInfo("CNC", "a16", 3, 11, 17, false, false, false, false, false, "Call on no carry", 0);
    private static final I8080OpInfo op0xd5 = new I8080OpInfo("PUSH", "D", 1, 11, 11, false, false, false, false, false, "Push DE on stack", 0);
    private static final I8080OpInfo op0xd6 = new I8080OpInfo("SUI", "d8", 2, 7, 7, true, true, true, true, true, "Subtract immediate from A", 0);
    private static final I8080OpInfo op0xd7 = new I8080OpInfo("RST", "2", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    private static final I8080OpInfo op0xd8 = new I8080OpInfo("RC", "", 1, 5, 11, false, false, false, false, false, "Return on carry", 0);
    private static final I8080OpInfo op0xd9 = new I8080OpInfo("*RET", "", 1, 10, 10, false, false, false, false, false, "Return", 0);
    private static final I8080OpInfo op0xda = new I8080OpInfo("JC", "a16", 3, 10, 10, false, false, false, false, false, "Jump on carry", 0);
    private static final I8080OpInfo op0xdb = new I8080OpInfo("IN", "d8", 2, 10, 10, false, false, false, false, false, "", 0);
    private static final I8080OpInfo op0xdc = new I8080OpInfo("CC", "a16", 3, 11, 17, false, false, false, false, false, "Call on carry", 0);
    private static final I8080OpInfo op0xdd = new I8080OpInfo("*CALL", "a16", 3, 17, 17, false, false, false, false, false, "Call unconditional", 0);
    private static final I8080OpInfo op0xde = new I8080OpInfo("SBI", "d8", 2, 7, 7, true, true, true, true, true, "Subtract immediate from A with borrow", 0);
    private static final I8080OpInfo op0xdf = new I8080OpInfo("RST", "3", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    // 0xe0-0xef
    private static final I8080OpInfo op0xe0 = new I8080OpInfo("RPO", "", 1, 5, 11, false, false, false, false, false, "Return on parity odd", 0);
    private static final I8080OpInfo op0xe1 = new I8080OpInfo("POP", "H", 1, 10, 10, false, false, false, false, false, "Pop HL off stack", 0);
    private static final I8080OpInfo op0xe2 = new I8080OpInfo("JPO", "a16", 3, 10, 10, false, false, false, false, false, "Jump on parity odd", 0);
    private static final I8080OpInfo op0xe3 = new I8080OpInfo("XTHL", "", 1, 18, 18, false, false, false, false, false, "Exchange top of stack and HL", 0);
    private static final I8080OpInfo op0xe4 = new I8080OpInfo("CPO", "a16", 3, 11, 17, false, false, false, false, false, "Call on parity odd", 0);
    private static final I8080OpInfo op0xe5 = new I8080OpInfo("PUSH", "H", 1, 11, 11, false, false, false, false, false, "Push HL on stack", 0);
    private static final I8080OpInfo op0xe6 = new I8080OpInfo("ANI", "d8", 2, 7, 7, true, true, false, true, true, "AND immediate with A", 0);
    private static final I8080OpInfo op0xe7 = new I8080OpInfo("RST", "4", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    private static final I8080OpInfo op0xe8 = new I8080OpInfo("RPE", "", 1, 5, 11, false, false, false, false, false, "Return on parity even", 0);
    private static final I8080OpInfo op0xe9 = new I8080OpInfo("PCHL", "", 1, 5, 5, false, false, false, false, false, "HL to program counter", 0);
    private static final I8080OpInfo op0xea = new I8080OpInfo("JPE", "a16", 3, 10, 10, false, false, false, false, false, "Jump on parity even", 0);
    private static final I8080OpInfo op0xeb = new I8080OpInfo("XCHG", "", 1, 4, 4, false, false, false, false, false, "Exchange DE and HL", 0);
    private static final I8080OpInfo op0xec = new I8080OpInfo("CPE", "a16", 3, 11, 17, false, false, false, false, false, "Call on parity even", 0);
    private static final I8080OpInfo op0xed = new I8080OpInfo("*CALL", "a16", 3, 17, 17, false, false, false, false, false, "Call unconditional", 0);
    private static final I8080OpInfo op0xee = new I8080OpInfo("XRI", "d8", 2, 7, 7, true, true, false, true, true, "Exclusive OR immediate with A", 0);
    private static final I8080OpInfo op0xef = new I8080OpInfo("RST", "5", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    // 0xf0-0xff
    private static final I8080OpInfo op0xf0 = new I8080OpInfo("RP", "", 1, 5, 11, false, false, false, false, false, "Return on positive", 0);
    private static final I8080OpInfo op0xf1 = new I8080OpInfo("POP", "PSW", 1, 10, 10, true, true, true, true, true, "Pop A and flags off stack", 0);
    private static final I8080OpInfo op0xf2 = new I8080OpInfo("JP", "a16", 3, 10, 10, false, false, false, false, false, "Jump on positive", 0);
    private static final I8080OpInfo op0xf3 = new I8080OpInfo("DI", "", 1, 4, 4, false, false, false, false, false, "", 0);
    private static final I8080OpInfo op0xf4 = new I8080OpInfo("CP", "a16", 3, 11, 17, false, false, false, false, false, "Call on positive", 0);
    private static final I8080OpInfo op0xf5 = new I8080OpInfo("PUSH", "PSW", 1, 11, 11, false, false, false, false, false, "Push A and flags on stack", 0);
    private static final I8080OpInfo op0xf6 = new I8080OpInfo("ORI", "d8", 2, 7, 7, true, true, false, true, true, "OR immediate with A", 0);
    private static final I8080OpInfo op0xf7 = new I8080OpInfo("RST", "6", 1, 11, 11, false, false, false, false, false, "Restart", 0);
    private static final I8080OpInfo op0xf8 = new I8080OpInfo("RM", "", 1, 5, 11, false, false, false, false, false, "Return on minus", 0);
    private static final I8080OpInfo op0xf9 = new I8080OpInfo("SPHL", "", 1, 5, 5, false, false, false, false, false, "HL to stack pointer", 0);
    private static final I8080OpInfo op0xfa = new I8080OpInfo("JM", "a16", 3, 10, 10, false, false, false, false, false, "Jump on minus", 0);
    private static final I8080OpInfo op0xfb = new I8080OpInfo("EI", "", 1, 4, 4, false, false, false, false, false, "", 0);
    private static final I8080OpInfo op0xfc = new I8080OpInfo("CM", "a16", 3, 11, 17, false, false, false, false, false, "Call on minus", 0);
    private static final I8080OpInfo op0xfd = new I8080OpInfo("*CALL", "a16", 3, 17, 17, false, false, false, false, false, "Call unconditional", 0);
    private static final I8080OpInfo op0xfe = new I8080OpInfo("CPI", "d8", 2, 7, 7, true, true, true, true, true, "Compare immediate with A", 0);
    private static final I8080OpInfo op0xff = new I8080OpInfo("RST", "7", 1, 11, 11, false, false, false, false, false, "Restart", 0);

    public static final I8080OpInfo[] opCode = new I8080OpInfo[] {
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

//    public static void main(String[] args) {
//        int[] flagSZP = new int[256];
//        for (int i=0; i<256; ++i) {
//            flagSZP[i] = (i & 0x80)
//                    | (i==0 ? 0x40 : 0)
//                    | (I8080.evenParity(i) ? 0x04 : 0)
//                    | 0x02;
//        }
//        for (int i = 0; i < flagSZP.length; i++) {
//            System.out.printf("%3d (%s) = %s\n", i, Util.toBinString(i, 8), Util.toBinString(flagSZP[i], 8));
//        }
//    }
//
//    public static final int[] flagsSZP = {
//        0x46, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06,
//        0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02,
//        0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02,
//        0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06,
//        0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02,
//        0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06,
//        0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06,
//        0x02, 0x06, 0x06, 0x02, 0x06, 0x02, 0x02, 0x06, 0x06, 0x02, 0x02, 0x06, 0x02, 0x06, 0x06, 0x02,
//        0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82,
//        0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86,
//        0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86,
//        0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82,
//        0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86,
//        0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82,
//        0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82,
//        0x86, 0x82, 0x82, 0x86, 0x82, 0x86, 0x86, 0x82, 0x82, 0x86, 0x86, 0x82, 0x86, 0x82, 0x82, 0x86
//    };
}
