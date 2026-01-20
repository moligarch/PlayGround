#pragma once
#ifndef DEVGUARD_STRING
#define DEVGUARD_STRING

#include <cstdint>
#include <string>
#include <type_traits>

#if defined(_MSC_VER)
#define DVGRD_FORCE_INLINE __forceinline
#else
#define DVGRD_FORCE_INLINE __attribute__((always_inline)) inline
#endif

namespace dvgrd::string {

    //============================================================
    // Compile-time entropy
    //============================================================
    namespace detail {

        constexpr uint64_t hash(const char* s) {
            uint64_t h = 1469598103934665603ull;
            while (*s) {
                h ^= static_cast<unsigned>(*s++);
                h *= 1099511628211ull;
            }
            return h;
        }

        constexpr uint64_t base_seed() {
            return hash(__TIME__) ^ hash(__DATE__);
        }

        template <typename CharT>
        constexpr CharT narrow(uint64_t v) {
            return static_cast<CharT>(v & ((1ull << (sizeof(CharT) * 8)) - 1));
        }

        //============================================================
        // Literal capture (structural NTTP)
        //============================================================
        template <typename CharT, std::size_t N>
        struct Literal {
            using value_type = CharT;
            static constexpr std::size_t size = N;
            CharT data[N];

            constexpr Literal(const CharT(&str)[N]) {
                for (std::size_t i = 0; i < N; ++i)
                    data[i] = str[i];
            }
        };
    } // namespace detail

    //============================================================
    // Obfuscated string
    //============================================================
    template <typename CharT, std::size_t N, uint64_t Salt>
    class SecureString {
    public:
        using value_type = CharT;

        constexpr explicit SecureString(const CharT(&str)[N])
            : key_(make_key()) {
            for (std::size_t i = 0; i < N; ++i)
                data_[i] = str[i] ^ stream(key_, i);
        }

        [[nodiscard]] DVGRD_FORCE_INLINE
            std::basic_string<CharT> decrypt() const {
            std::basic_string<CharT> out;
            out.resize(N);
            for (std::size_t i = 0; i < N; ++i)
                out[i] = data_[i] ^ stream(key_, i);
            return out;
        }

        constexpr std::size_t size() const noexcept { return N; }

    private:
        static constexpr CharT make_key() {
            uint64_t k =
                detail::base_seed()
                ^ Salt
                ^ (N * 0x9E3779B97F4A7C15ull);
            return detail::narrow<CharT>(k);
        }

        static constexpr CharT stream(CharT key, std::size_t i) {
            return static_cast<CharT>(key + static_cast<CharT>(i * 131));
        }

        CharT data_[N]{};
        const CharT key_;
    };

#if defined(_MSVC_LANG) && (_MSVC_LANG < 202002L) // C++ >= 20
    template <typename CharT, std::size_t N>
    constexpr auto make_obf_raw(const CharT (&str)[N]) {
        constexpr uint64_t salt =
            detail::hash(__FILE__)
            ^ __LINE__
            ^ __COUNTER__;

        return SecureString<CharT, N, salt>{ str };
    }
#endif

    //============================================================
    // Literal operators (char / wchar_t / char16_t / char32_t)
    //============================================================
#if defined(_MSVC_LANG) && (_MSVC_LANG >= 202002L) // C++ >= 20

    template <detail::Literal Str>
    constexpr auto operator""_enc() {
        using CharT = typename decltype(Str)::value_type;
        constexpr std::size_t N = decltype(Str)::size; // drop null
        constexpr uint64_t salt =
            detail::hash(__FILE__)
            ^ __LINE__
            ^ __COUNTER__;

        return SecureString<CharT, N, salt>{ Str.data };
    }

    template <detail::Literal Str>
    DVGRD_FORCE_INLINE auto operator""_sec() {
        return operator""_enc<Str>().decrypt();
    }

#endif // C++ >= 20

} // namespace dvgrd::string


#if defined(_MSVC_LANG) && (_MSVC_LANG >= 202002L) // C++ >= 20

using dvgrd::string::operator""_enc;
using dvgrd::string::operator""_sec;
#define DVGRD_PROTECT_STR(str) str##_sec

#else // C++ < 20

#define DVGRD_PROTECT_STR(str) (::dvgrd::string::make_obf_raw(str).decrypt())

#endif

#endif // DEVGUARD_STRING