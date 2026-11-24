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
package i8080.gui;

/**
 *
 * @author Barend Scholtus
 */
public class Util {

    /**
     *
     * @param val
     * @param minLength
     * @return
     */
    public static String toHexString(int val, int minLength) {
        StringBuilder sb = new StringBuilder();
        while (val != 0) {
            switch (val & 0xf) {
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
            val >>>= 4;
        }
        while (sb.length() < minLength) {
            sb.append('0');
        }
        return sb.reverse().toString();
    }

    /**
     *
     * @param val
     * @param minLength
     * @return
     */
    public static String toBinString(int val, int minLength) {
        StringBuilder sb = new StringBuilder(32);
        while (val != 0) {
            sb.append((val & 0x1) == 0 ? '0' : '1');
            val >>>= 1;
        }
        while (sb.length() < minLength) {
            sb.append('0');
        }
        return sb.reverse().toString();
    }
}
