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

import java.util.Formatter;

/**
 *
 * @author Barend Scholtus
 */
public class I8080 {

    int B;
    int C;
    int D;
    int E;
    int H;
    int L;
    int A;
    int PC;
    int SP;
    boolean Carry;
    boolean Zero;
    boolean Sign;
    boolean Parity;
    boolean AuxCarry;
    private I8080Context ctx;

    public I8080(I8080Context env) {
        this.ctx = env;
//        reset();
    }

    public void reset() {
        PC = 0;
    }

    public int AF() {
        return (A << 8) | getConditionBits();
    }

    public int BC() {
        return (B << 8) | C;
    }

    public int DE() {
        return (D << 8) | E;
    }

    public int HL() {
        return (H << 8) | L;
    }

    private int readOprnd() {
        return ctx.read(PC++);
    }

    private int readOprnd2() {
        int t = ctx.read2(PC);
        PC += 2;
        return t;
    }

    private void write(int addr, int data) {
        ctx.write(addr, data);
    }

    public int getConditionBits() {
        return (Carry ? 0x01 : 0)
                | (0x02)
                | (Parity ? 0x04 : 0)
                | (AuxCarry ? 0x10 : 0)
                | (Zero ? 0x40 : 0)
                | (Sign ? 0x80 : 0);
    }

    private void setConditionBits(int t) {
        Carry = (t & 0x01) != 0;
        Parity = (t & 0x04) != 0;
        AuxCarry = (t & 0x10) != 0;
        Zero = (t & 0x40) != 0;
        Sign = (t & 0x80) != 0;
    }

    public int instruction() {
        int t;
        int opcode = ctx.read(PC++);
        switch (opcode) {
            case 0x00:
                break;
            case 0x01:
                C = ctx.read(PC++);
                B = ctx.read(PC++);
                break;
            case 0x02:
                write(BC(), A);
                break;
            case 0x03:
                if (++C > 0xFF) {
                    C = 0;
                    if (++B > 0xFF) {
                        B = 0;
                    }
                }
                break;
            case 0x04:
                AuxCarry = (B & 0xF) == 0xF;
                if (++B > 0xFF) {
                    B = 0;
                }
                Sign = (B & 0x80) != 0;
                Zero = B == 0;
                Parity = evenParity(B);
                break;
            case 0x05:
                AuxCarry = (B & 0xF) == 0;
                if (--B < 0) {
                    B = 0xFF;
                }
                Sign = (B & 0x80) != 0;
                Zero = B == 0;
                Parity = evenParity(B);
                break;
            case 0x06:
                B = ctx.read(PC++);
                break;
            case 0x07:
                break;
            case 0x08:
                break;
            case 0x09:
                break;
            case 0x0A:
                ctx.read(BC());
                break;
            case 0x0B:
                if (--C < 0) {
                    C = 0xFF;
                    if (--B < 0) {
                        B = 0xFF;
                    }
                }
                break;
            case 0x0C:
                AuxCarry = (C & 0xF) == 0xF;
                if (++C > 0xFF) {
                    C = 0;
                }
                Sign = (C & 0x80) != 0;
                Zero = C == 0;
                Parity = evenParity(C);
                break;
            case 0x0D:
                AuxCarry = (C & 0xF) == 0;
                if (--C < 0) {
                    C = 0xFF;
                }
                Sign = (C & 0x80) != 0;
                Zero = C == 0;
                Parity = evenParity(C);
                break;
            case 0x0E:
                C = ctx.read(PC++);
                break;
            case 0x0F:
                break;
            case 0x10:
                break;
            case 0x11:
                E = ctx.read(PC++);
                D = ctx.read(PC++);
                break;
            case 0x12:
                write(DE(), A);
                break;
            case 0x13:
                if (++E > 0xFF) {
                    E = 0;
                    if (++D > 0xFF) {
                        D = 0;
                    }
                }
                break;
            case 0x14:
                AuxCarry = (D & 0xF) == 0xF;
                if (++D > 0xFF) {
                    D = 0;
                }
                Sign = (D & 0x80) != 0;
                Zero = D == 0;
                Parity = evenParity(D);
                break;
            case 0x15:
                AuxCarry = (D & 0xF) == 0;
                if (--D < 0) {
                    D = 0xFF;
                }
                Sign = (D & 0x80) != 0;
                Zero = D == 0;
                Parity = evenParity(D);
                break;
            case 0x16:
                D = ctx.read(PC++);
                break;
            case 0x17:
                break;
            case 0x18:
                break;
            case 0x19:
                break;
            case 0x1A:
                ctx.read(DE());
                break;
            case 0x1B:
                if (--E < 0) {
                    E = 0xFF;
                    if (--D < 0) {
                        D = 0xFF;
                    }
                }
                break;
            case 0x1C:
                AuxCarry = (E & 0xF) == 0xF;
                if (++E > 0xFF) {
                    E = 0;
                }
                Sign = (E & 0x80) != 0;
                Zero = E == 0;
                Parity = evenParity(E);
                break;
            case 0x1D:
                AuxCarry = (E & 0xF) == 0;
                if (--E < 0) {
                    E = 0xFF;
                }
                Sign = (E & 0x80) != 0;
                Zero = E == 0;
                Parity = evenParity(E);
                break;
            case 0x1E:
                E = ctx.read(PC++);
                break;
            case 0x1F:
                break;
            case 0x20:
                break;
            case 0x21:
                L = ctx.read(PC++);
                H = ctx.read(PC++);
                break;
            case 0x22:
                t = readOprnd2();
                write(t++, L);
                write(t, H);
                break;
            case 0x23:
                if (++L > 0xFF) {
                    L = 0;
                    if (++H > 0xFF) {
                        H = 0;
                    }
                }
                break;
            case 0x24:
                AuxCarry = (H & 0xF) == 0xF;
                if (++H > 0xFF) {
                    H = 0;
                }
                Sign = (H & 0x80) != 0;
                Zero = H == 0;
                Parity = evenParity(H);
                break;
            case 0x25:
                AuxCarry = (H & 0xF) == 0;
                if (--H < 0) {
                    H = 0xFF;
                }
                Sign = (H & 0x80) != 0;
                Zero = H == 0;
                Parity = evenParity(H);
                break;
            case 0x26:
                H = ctx.read(PC++);
                break;
            case 0x27:
                break;
            case 0x28:
                break;
            case 0x29:
                break;
            case 0x2A:
                t = readOprnd2();
                L = ctx.read(t++);
                H = ctx.read(t);
                break;
            case 0x2B:
                if (--L < 0) {
                    L = 0xFF;
                    if (--H < 0) {
                        H = 0xFF;
                    }
                }
                break;
            case 0x2C:
                AuxCarry = (L & 0xF) == 0xF;
                if (++L > 0xFF) {
                    L = 0;
                }
                Sign = (L & 0x80) != 0;
                Zero = L == 0;
                Parity = evenParity(L);
                break;
            case 0x2D:
                AuxCarry = (L & 0xF) == 0;
                if (--L < 0) {
                    L = 0xFF;
                }
                Sign = (L & 0x80) != 0;
                Zero = L == 0;
                Parity = evenParity(L);
                break;
            case 0x2E:
                L = ctx.read(PC++);
                break;
            case 0x2F:
                break;
            case 0x30:
                break;
            case 0x31:
                SP = readOprnd2();
                break;
            case 0x32:
                write(readOprnd2(), A);
                break;
            case 0x33:
                if (++SP > 0xFFFF) {
                    SP = 0;
                }
                break;
            case 0x34:
                t = ctx.read(HL());
                AuxCarry = (t & 0xF) == 0xF;
                if (++t > 0xFF) {
                    t = 0;
                }
                Sign = (t & 0x80) != 0;
                Zero = t == 0;
                Parity = evenParity(t);
                write(HL(), t);
                break;
            case 0x35:
                t = ctx.read(HL());
                AuxCarry = (t & 0xF) == 0;
                if (--t < 0) {
                    t = 0xFF;
                }
                Sign = (t & 0x80) != 0;
                Zero = t == 0;
                Parity = evenParity(t);
                write(HL(), t);
                break;
            case 0x36:
                write(HL(), ctx.read(PC++));
                break;
            case 0x37:
                break;
            case 0x38:
                break;
            case 0x39:
                break;
            case 0x3A:
                A = ctx.read(readOprnd2());
                break;
            case 0x3B:
                if (--SP < 0) {
                    SP = 0xFFFF;
                }
                break;
            case 0x3C:
                AuxCarry = (A & 0xF) == 0xF;
                if (++A > 0xFF) {
                    A = 0;
                }
                Sign = (A & 0x80) != 0;
                Zero = A == 0;
                Parity = evenParity(A);
                break;
            case 0x3D:
                AuxCarry = (A & 0xF) == 0;
                if (--A < 0) {
                    A = 0xFF;
                }
                Sign = (A & 0x80) != 0;
                Zero = A == 0;
                Parity = evenParity(A);
                break;
            case 0x3E:
                A = ctx.read(PC++);
                break;
            case 0x3F:
                break;
            case 0x40: /*B = B;*/ break;
            case 0x41:
                B = C;
                break;
            case 0x42:
                B = D;
                break;
            case 0x43:
                B = E;
                break;
            case 0x44:
                B = H;
                break;
            case 0x45:
                B = L;
                break;
            case 0x46:
                B = ctx.read(HL());
                break;
            case 0x47:
                B = A;
                break;
            case 0x48:
                C = B;
                break;
            case 0x49: /*C = C;*/ break;
            case 0x4A:
                C = D;
                break;
            case 0x4B:
                C = E;
                break;
            case 0x4C:
                C = H;
                break;
            case 0x4D:
                C = L;
                break;
            case 0x4E:
                C = ctx.read(HL());
                break;
            case 0x4F:
                C = A;
                break;
            case 0x50:
                D = B;
                break;
            case 0x51:
                D = C;
                break;
            case 0x52: /*D = D;*/ break;
            case 0x53:
                D = E;
                break;
            case 0x54:
                D = H;
                break;
            case 0x55:
                D = L;
                break;
            case 0x56:
                D = ctx.read(HL());
                break;
            case 0x57:
                D = A;
                break;
            case 0x58:
                E = B;
                break;
            case 0x59:
                E = C;
                break;
            case 0x5A:
                E = D;
                break;
            case 0x5B: /*E = E;*/ break;
            case 0x5C:
                E = H;
                break;
            case 0x5D:
                E = L;
                break;
            case 0x5E:
                E = ctx.read(HL());
                break;
            case 0x5F:
                E = A;
                break;
            case 0x60:
                H = B;
                break;
            case 0x61:
                H = C;
                break;
            case 0x62:
                H = D;
                break;
            case 0x63:
                H = E;
                break;
            case 0x64: /*H = H;*/ break;
            case 0x65:
                H = L;
                break;
            case 0x66:
                H = ctx.read(HL());
                break;
            case 0x67:
                H = A;
                break;
            case 0x68:
                L = B;
                break;
            case 0x69:
                L = C;
                break;
            case 0x6A:
                L = D;
                break;
            case 0x6B:
                L = E;
                break;
            case 0x6C:
                L = H;
                break;
            case 0x6D: /*L = L;*/ break;
            case 0x6E:
                L = ctx.read(HL());
                break;
            case 0x6F:
                L = A;
                break;
            case 0x70:
                write(HL(), B);
                break;
            case 0x71:
                write(HL(), C);
                break;
            case 0x72:
                write(HL(), D);
                break;
            case 0x73:
                write(HL(), E);
                break;
            case 0x74:
                write(HL(), H);
                break;
            case 0x75:
                write(HL(), L);
                break;
            case 0x76:
                break;
            case 0x77:
                write(HL(), A);
                break;
            case 0x78:
                A = B;
                break;
            case 0x79:
                A = C;
                break;
            case 0x7A:
                A = D;
                break;
            case 0x7B:
                A = E;
                break;
            case 0x7C:
                A = H;
                break;
            case 0x7D:
                A = L;
                break;
            case 0x7E:
                A = ctx.read(HL());
                break;
            case 0x7F: /*A = A;*/ break;
            case 0x80:
                break;
            case 0x81:
                break;
            case 0x82:
                break;
            case 0x83:
                break;
            case 0x84:
                break;
            case 0x85:
                break;
            case 0x86:
                break;
            case 0x87:
                break;
            case 0x88:
                break;
            case 0x89:
                break;
            case 0x8A:
                break;
            case 0x8B:
                break;
            case 0x8C:
                break;
            case 0x8D:
                break;
            case 0x8E:
                break;
            case 0x8F:
                break;
            case 0x90:
                break;
            case 0x91:
                break;
            case 0x92:
                break;
            case 0x93:
                break;
            case 0x94:
                break;
            case 0x95:
                break;
            case 0x96:
                break;
            case 0x97:
                break;
            case 0x98:
                break;
            case 0x99:
                break;
            case 0x9A:
                break;
            case 0x9B:
                break;
            case 0x9C:
                break;
            case 0x9D:
                break;
            case 0x9E:
                break;
            case 0x9F:
                break;
            case 0xA0:
                break;
            case 0xA1:
                break;
            case 0xA2:
                break;
            case 0xA3:
                break;
            case 0xA4:
                break;
            case 0xA5:
                break;
            case 0xA6:
                break;
            case 0xA7:
                break;
            case 0xA8:
                break;
            case 0xA9:
                break;
            case 0xAA:
                break;
            case 0xAB:
                break;
            case 0xAC:
                break;
            case 0xAD:
                break;
            case 0xAE:
                break;
            case 0xAF:
                break;
            case 0xB0:
                break;
            case 0xB1:
                break;
            case 0xB2:
                break;
            case 0xB3:
                break;
            case 0xB4:
                break;
            case 0xB5:
                break;
            case 0xB6:
                break;
            case 0xB7:
                break;
            case 0xB8:
                break;
            case 0xB9:
                break;
            case 0xBA:
                break;
            case 0xBB:
                break;
            case 0xBC:
                break;
            case 0xBD:
                break;
            case 0xBE:
                break;
            case 0xBF:
                break;
            case 0xC0:
                return opReturn(!Zero); // RNZ - return if not zero
            case 0xC1:
                C = ctx.read(SP++);
                B = ctx.read(SP++);
                break;
            case 0xC2:
                PC = Zero ? PC + 2 : ctx.read2(PC);
                break; // JNZ - jump if not zero
            case 0xC3:
                PC = ctx.read2(PC);
                break; // JMP
            case 0xC4:
                return opCall(!Zero); // CNZ - call if not zero
            case 0xC5:
                write(--SP, B);
                write(--SP, C);
                break;
            case 0xC6:
                break;
            case 0xC7:
                break;
            case 0xC8:
                return opReturn(Zero); // RZ - return if zero
            case 0xC9:
                PC = ctx.read2(SP);
                SP += 2;
                break; // RET
            case 0xCA:
                PC = Zero ? ctx.read2(PC) : PC + 2;
                break; // JZ - jump if zero
            case 0xCB:
                break;
            case 0xCC:
                return opCall(Zero); // CZ - call if zero
            case 0xCD:
                return opCall(true); // CALL
            case 0xCE:
                break;
            case 0xCF:
                break;
            case 0xD0:
                return opReturn(!Carry); // RNC - return if not carry
            case 0xD1:
                E = ctx.read(SP++);
                D = ctx.read(SP++);
                break;
            case 0xD2:
                PC = Carry ? PC + 2 : ctx.read2(PC);
                break; // JNC - jump if not carry
            case 0xD3:
                break;
            case 0xD4:
                return opCall(!Carry); // CNC - call if not Carry
            case 0xD5:
                write(--SP, D);
                write(--SP, E);
                break;
            case 0xD6:
                break;
            case 0xD7:
                break;
            case 0xD8:
                return opReturn(Carry); // RC - return if carry
            case 0xD9:
                break;
            case 0xDA:
                PC = Carry ? ctx.read2(PC) : PC + 2;
                break; // JC - jump if carry
            case 0xDB:
                break;
            case 0xDC:
                return opCall(Carry); // CC - call if Carry
            case 0xDD:
                break;
            case 0xDE:
                break;
            case 0xDF:
                break;
            case 0xE0:
                return opReturn(!Parity); // RPO - return if parity odd
            case 0xE1:
                L = ctx.read(SP++);
                H = ctx.read(SP++);
                break;
            case 0xE2:
                PC = Parity ? PC + 2 : ctx.read2(PC);
                break; // JPO - jump if parity odd
            case 0xE3:
                t = L;
                L = ctx.read(SP);
                write(SP, t);
                t = H;
                H = ctx.read(SP + 1);
                write(SP + 1, t);
                break;
            case 0xE4:
                return opCall(!Parity); // CPO - call if Parity odd
            case 0xE5:
                write(--SP, H);
                write(--SP, L);
                break;
            case 0xE6:
                break;
            case 0xE7:
                break;
            case 0xE8:
                return opReturn(Parity); // RPE - return if parity even
            case 0xE9:
                PC = (H << 8) | L;
                break; // PCHL - load PC from HL
            case 0xEA:
                PC = Parity ? ctx.read2(PC) : PC + 2;
                break; // JPE - jump if parity even
            case 0xEB:
                t = H;
                H = L;
                L = t;
                t = E;
                E = D;
                D = t;
                break;
            case 0xEC:
                return opCall(Parity); // CPE - call if Parity even
            case 0xED:
                break;
            case 0xEE:
                break;
            case 0xEF:
                break;
            case 0xF0:
                return opReturn(!Sign); // RP - return if plus
            case 0xF1:
                setConditionBits(ctx.read(SP++));
                A = ctx.read(SP++);
                break;
            case 0xF2:
                PC = Sign ? PC + 2 : ctx.read2(PC);
                break; // JP - jump if positive
            case 0xF3:
                break;
            case 0xF4:
                return opCall(!Sign); // CP - call if plus
            case 0xF5:
                write(--SP, A);
                write(--SP, getConditionBits());
                break;
            case 0xF6:
                break;
            case 0xF7:
                break;
            case 0xF8:
                return opReturn(Sign); // RM - return if minus
            case 0xF9:
                SP = (H << 8) | L;
                break; // SPHL - load SP from HL
            case 0xFA:
                PC = Sign ? ctx.read2(PC) : PC + 2;
                break; // JM - jump if minus
            case 0xFB:
                break;
            case 0xFC:
                return opCall(Sign); // CM - call if minus
            case 0xFD:
                break;
            case 0xFE:
                break;
            case 0xFF:
                break;
            default:
                throw new RuntimeException("opcode " + opcode + " does not exist");
        }

        return I8080OpInfo.opCode[opcode].cyclesShort;
    }

    private int opCall(boolean cond) {
        if (cond) {
            write(--SP, (PC + 2) >>> 8);
            write(--SP, (PC + 2) & 0xFF);
            PC = ctx.read2(PC);
            return 17;
        } else {
            PC += 2;
            return 11;
        }
    }

    private int opReturn(boolean cond) {
        if (cond) {
            PC = ctx.read2(SP);
            SP += 2;
            return 11;
        } else {
            return 5;
        }
    }

//    public static final String[] opMnemonic = new String[]{
//        "RNZ", "POP", "JNZ", "JMP", "CNZ", "PUSH", "", "", // 0xC?
//        "RZ", "RET", "JZ", "", "CZ", "CALL", "", "",
//        "RNC", "POP", "JNC", "", "CNC", "PUSH", "", "", // 0xD?
//        "RC", "", "JC", "", "CC", "", "", "",
//        "RPO", "POP", "JPO", "XTHL", "CPO", "PUSH", "", "", // 0xE?
//        "RPE", "PCHL", "JPE", "XCHG", "CPE", "", "", "",
//        "RP", "POP", "JP", "", "CP", "PUSH", "", "", // 0xF?
//        "RM", "SPHL", "JM", "", "CM", "", "", "",};
//
//    public String disasm(int addr) {
//        return "";
//    }

    public int PC() {
        return PC;
    }

    public int SP() {
        return SP;
    }

    public boolean isCarry() {
        return Carry;
    }

    public boolean isParity() {
        return Parity;
    }

    public boolean isAuxCarry() {
        return AuxCarry;
    }

    public boolean isZero() {
        return Zero;
    }

    public boolean isSign() {
        return Sign;
    }

    /**
     * Returns whether val has even parity or not.
     * Based on http://graphics.stanford.edu/~seander/bithacks.html
     * @param val the value for which to determine whether it has even parity
     * @return true if val has even parity
     */
    public static boolean evenParity(int val) {
        return (0x6996 >> ((val ^ (val >> 4)) & 0xF) & 1) == 0;
    }
}
